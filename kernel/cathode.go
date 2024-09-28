//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/fogfish/swarm"
)

type Cathode interface {
	Ack(ctx context.Context, digest string) error
	Err(ctx context.Context, digest string, err error) error
	Ask(ctx context.Context) ([]swarm.Bag, error)
}

// Decode message from wire format
type Decoder[T any] interface{ Decode([]byte) (T, error) }

type Router = interface {
	Route(context.Context, swarm.Bag) error
}

type Dequeuer struct {
	sync.WaitGroup
	sync.RWMutex

	// Control-plane stop channel used by go routines to stop I/O on data channels
	context context.Context
	cancel  context.CancelFunc

	// Kernel configuration
	Config swarm.Config

	// event router, binds category with destination channel
	router map[string]Router

	// Cathode is the reader port on message broker
	Cathode Cathode
}

func NewDequeuer(cathode Cathode, config swarm.Config) *Dequeuer {
	ctx, can := context.WithCancel(context.Background())

	return &Dequeuer{
		Config:  config,
		context: ctx,
		cancel:  can,
		router:  make(map[string]Router),
		Cathode: cathode,
	}
}

// Close enqueuer
func (k *Dequeuer) Close() {
	k.cancel()
	k.WaitGroup.Wait()
}

func (k *Dequeuer) Await() {
	if spawner, ok := k.Cathode.(interface{ Run() }); ok {
		go spawner.Run()
	}

	k.receive()
	<-k.context.Done()
	k.WaitGroup.Wait()
}

// internal infinite receive loop.
// waiting for message from event buses and queues and schedules it for delivery.
func (k *Dequeuer) receive() {
	asker := func() {
		seq, err := k.Cathode.Ask(k.context)
		if k.Config.StdErr != nil && err != nil {
			k.Config.StdErr <- err
			return
		}

		for i := 0; i < len(seq); i++ {
			bag := seq[i]

			k.RWMutex.RLock()
			r, has := k.router[bag.Ctx.Category]
			k.RWMutex.RUnlock()

			if has {
				err := r.Route(k.context, bag)
				if k.Config.StdErr != nil && err != nil {
					k.Config.StdErr <- err
					return
				}
			}
		}
	}

	k.WaitGroup.Add(1)
	go func() {
		slog.Debug("kernel receive loop started")

	exit:
		for {
			select {
			case <-k.context.Done():
				break exit
			default:
			}

			select {
			case <-k.context.Done():
				break exit
			case <-time.After(k.Config.PollFrequency):
				asker()
			}
		}

		k.WaitGroup.Done()
		slog.Debug("kernel receive loop stopped")
	}()
}

// Dequeue creates pair of channels within kernel to enqueue messages
func Dequeue[T any](k *Dequeuer, cat string, codec Decoder[T]) ( /*rcv*/ <-chan swarm.Msg[T] /*ack*/, chan<- swarm.Msg[T]) {
	rcv := make(chan swarm.Msg[T], k.Config.CapRcv)
	ack := make(chan swarm.Msg[T], k.Config.CapAck)

	k.RWMutex.Lock()
	k.router[cat] = router[T]{ch: rcv, codec: codec}
	k.RWMutex.Unlock()

	// emitter routine
	acks := func(msg swarm.Msg[T]) {
		if msg.Ctx.Error == nil {
			err := k.Cathode.Ack(k.context, msg.Ctx.Digest)
			if k.Config.StdErr != nil && err != nil {
				k.Config.StdErr <- err
			}
		} else {
			err := k.Cathode.Err(k.context, msg.Ctx.Digest, msg.Ctx.Error)
			if k.Config.StdErr != nil && err != nil {
				k.Config.StdErr <- err
			}
		}
	}

	k.WaitGroup.Add(1)
	go func() {
		slog.Debug("kernel dequeue started", "cat", cat)

	exit:
		for {
			// The try-receive operation here is to
			// try to exit the sender goroutine as
			// early as possible. Try-receive and
			// try-send select blocks are specially
			// optimized by the standard Go
			// compiler, so they are very efficient.
			select {
			case <-k.context.Done():
				break exit
			default:
			}

			select {
			case <-k.context.Done():
				break exit
			case msg := <-ack:
				acks(msg)
			}
		}

		backlog := len(ack)
		close(ack)

		if backlog != 0 {
			for msg := range ack {
				acks(msg)
			}
		}

		k.WaitGroup.Done()
		slog.Debug("kernel dequeue stopped", "cat", cat)
	}()

	return rcv, ack
}