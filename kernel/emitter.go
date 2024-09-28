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
	"sync"

	"github.com/fogfish/swarm"
)

// Emitter defines on-the-wire protocol for [swarm.Bag]
type Emitter interface {
	Enq(context.Context, swarm.Bag) error
}

// Encodes message into wire format
type Encoder[T any] interface{ Encode(T) ([]byte, error) }

// Messaging Egress port
type Enqueuer struct {
	sync.WaitGroup

	// Control-plane stop channel used by go routines to stop I/O on data channels
	context context.Context
	cancel  context.CancelFunc

	// Kernel configuration
	Config swarm.Config

	// Emitter is the writer port on message broker
	Emitter Emitter
}

func NewEnqueuer(emitter Emitter, config swarm.Config) *Enqueuer {
	ctx, can := context.WithCancel(context.Background())

	return &Enqueuer{
		Config:  config,
		context: ctx,
		cancel:  can,
		Emitter: emitter,
	}
}

// Close enqueuer
func (k *Enqueuer) Close() {
	k.cancel()
	k.WaitGroup.Wait()
}

// Await enqueue
func (k *Enqueuer) Await() {
	<-k.context.Done()
	k.WaitGroup.Wait()
}

// Enqueue creates pair of channels within kernel to enqueue messages
func Enqueue[T any](k *Enqueuer, cat string, codec Encoder[T]) ( /*snd*/ chan<- T /*dlq*/, <-chan T) {
	snd := make(chan T, k.Config.CapOut)
	dlq := make(chan T, k.Config.CapDLQ)

	// emitter routine
	emit := func(obj T) {
		msg, err := codec.Encode(obj)
		if err != nil {
			dlq <- obj
			if k.Config.StdErr != nil {
				k.Config.StdErr <- err
			}
			return
		}

		ctx := swarm.NewContext(context.Background(), cat, "")
		bag := swarm.Bag{Ctx: ctx, Object: msg}

		if err := k.Emitter.Enq(context.Background(), bag); err != nil {
			dlq <- obj
			if k.Config.StdErr != nil {
				k.Config.StdErr <- err
			}
			return
		}
	}

	k.WaitGroup.Add(1)
	go func() {
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
			case obj := <-snd:
				emit(obj)
			}
		}

		backlog := len(snd)
		close(snd)

		if backlog != 0 {
			for obj := range snd {
				emit(obj)
			}
		}

		k.WaitGroup.Done()
	}()

	return snd, dlq
}