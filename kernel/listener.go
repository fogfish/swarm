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

// Listener defines on-the-wire protocol for [swarm.Bag], covering the ingress use-cases.
type Listener interface {
	Ack(ctx context.Context, digest string) error
	Err(ctx context.Context, digest string, err error) error
	Ask(ctx context.Context) ([]swarm.Bag, error)
}

// Decode message from wire format
type Decoder[T any] interface {
	Category() string
	Decode([]byte) (T, error)
}

// Routes messages from the ingress to the destination channel.
type Router = interface {
	Route(context.Context, swarm.Bag) error
}

type ListenerKernel struct {
	sync.WaitGroup
	sync.RWMutex

	// Control-plane stop channel used by go routines to stop I/O on data channels
	context context.Context
	cancel  context.CancelFunc

	// Kernel configuration
	Config swarm.Config

	// event router, binds category with destination channel
	router map[string]Router

	// Listener is the reader port on message broker
	Listener Listener
}

func NewListener(listener Listener, config swarm.Config) *ListenerKernel {
	return builder().Dequeuer(listener, config)
}

// Creates instance of broker reader
func newListener(listener Listener, config swarm.Config) *ListenerKernel {
	ctx, can := context.WithCancel(context.Background())

	// Must not be 0
	if config.PollerPool == 0 {
		config.PollerPool = 1
	}

	return &ListenerKernel{
		Config:   config,
		context:  ctx,
		cancel:   can,
		router:   make(map[string]Router),
		Listener: listener,
	}
}

// Closes broker reader, gracefully shutdowns all I/O
func (k *ListenerKernel) Close() {
	k.cancel()
	k.WaitGroup.Wait()
}

// Await reader to complete
func (k *ListenerKernel) Await() {
	if spawner, ok := k.Listener.(interface{ Run(context.Context) }); ok {
		go spawner.Run(k.context)
	}

	k.receive()

	<-k.context.Done()
	k.WaitGroup.Wait()
}

// internal infinite receive loop.
// waiting for message from event buses and queues and schedules it for delivery.
func (k *ListenerKernel) receive() {
	asker := func() {
		var seq []swarm.Bag
		err := k.Config.Backoff.Retry(
			func() (exx error) {
				seq, exx = k.Listener.Ask(k.context)
				return
			},
		)
		if k.Config.StdErr != nil && err != nil {
			k.Config.StdErr <- swarm.ErrDequeue.With(err)
			return
		}

		for i := 0; i < len(seq); i++ {
			bag := seq[i]

			k.RWMutex.RLock()
			r, has := k.router[bag.Category]
			k.RWMutex.RUnlock()

			if has {
				err := r.Route(k.context, bag)
				if k.Config.StdErr != nil && err != nil {
					k.Config.StdErr <- swarm.ErrDequeue.With(err)
					return
				}
			} else {
				slog.Warn("Unknown category",
					slog.Any("cat", bag.Category),
					slog.Any("kernel", k.Config.Agent),
				)
				slog.Debug("Unknown category",
					slog.Any("cat", bag.Category),
					slog.Any("kernel", k.Config.Agent),
					slog.Any("bag", bag),
				)
				if k.Config.FailOnUnknownCategory {
					k.Listener.Err(k.context, bag.Digest, swarm.ErrDequeue.With(swarm.ErrCatUnknown.With(nil, bag.Category)))
				}
			}
		}
	}

	for pid := 0; pid < k.Config.PollerPool; pid++ {
		k.WaitGroup.Add(1)
		go func() {
			slog.Debug("kernel poller started", "pid", pid)
			defer slog.Debug("kernel poller stopped", "pid", pid)

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
		}()
	}
}

// RecvChan creates pair of channels within kernel to enqueue messages
func RecvChan[T any](k *ListenerKernel, cat string, codec Decoder[T]) ( /*rcv*/ <-chan swarm.Msg[T] /*ack*/, chan<- swarm.Msg[T]) {
	rcv := make(chan swarm.Msg[T], k.Config.CapRcv)
	ack := make(chan swarm.Msg[T], k.Config.CapAck)

	k.RWMutex.Lock()
	k.router[cat] = router[T]{ch: rcv, codec: codec}
	k.RWMutex.Unlock()

	// emitter routine
	acks := func(msg swarm.Msg[T]) {
		if msg.Error == nil {
			err := k.Config.Backoff.Retry(
				func() error {
					return k.Listener.Ack(k.context, msg.Digest)
				},
			)
			if k.Config.StdErr != nil && err != nil {
				k.Config.StdErr <- swarm.ErrDequeue.With(err)
			}
		} else {
			err := k.Config.Backoff.Retry(
				func() error {
					return k.Listener.Err(k.context, msg.Digest, msg.Error)
				},
			)
			if k.Config.StdErr != nil && err != nil {
				k.Config.StdErr <- swarm.ErrDequeue.With(err)
			}
		}
	}

	k.WaitGroup.Add(1)
	go func() {
		slog.Debug("kernel dequeue started", "cat", cat)
		defer slog.Debug("kernel dequeue stopped", "cat", cat)

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
	}()

	return rcv, ack
}
