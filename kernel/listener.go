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

	"github.com/fogfish/golem/optics"
	"github.com/fogfish/swarm"
)

// Listener defines on-the-wire protocol for [swarm.Bag], covering the ingress use-cases.
type Listener interface {
	Ack(ctx context.Context, digest swarm.Digest) error
	Err(ctx context.Context, digest swarm.Digest, err error) error
	Ask(ctx context.Context) ([]swarm.Bag, error)
	Close() error
}

// Decode message from wire format
type Decoder[T any] interface {
	Category() string
	Decode(swarm.Bag) (T, error)
}

// Routes messages from the ingress to the destination channel.
type Router = interface {
	Route(context.Context, swarm.Bag) error
}

// The ingress part of the kernel is used to dequeue messages from message broker.
type ListenerIO struct {
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

func NewListener(listener Listener, config swarm.Config) *ListenerIO {
	return builder().Listener(listener, config)
}

// Creates instance of broker reader
func newListener(listener Listener, config swarm.Config) *ListenerIO {
	ctx, can := context.WithCancel(context.Background())

	// Must not be 0
	if config.PollerPool == 0 {
		config.PollerPool = 1
	}

	return &ListenerIO{
		Config:   config,
		context:  ctx,
		cancel:   can,
		router:   make(map[string]Router),
		Listener: listener,
	}
}

// Closes broker reader, gracefully shutdowns all I/O
func (k *ListenerIO) Close() {
	k.cancel()
	k.WaitGroup.Wait()
	k.Listener.Close()
}

// Await reader to complete
func (k *ListenerIO) Await() {
	if spawner, ok := k.Listener.(interface{ Run(context.Context) }); ok {
		go spawner.Run(k.context)
	}

	k.receive()

	<-k.context.Done()
	k.WaitGroup.Wait()
	k.Listener.Close()
}

// internal infinite receive loop.
// waiting for message from event buses and queues and schedules it for delivery.
func (k *ListenerIO) receive() {
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

// RecvChan creates pair of channels within kernel to receive messages
func RecvChan[T any](k *ListenerIO, codec Decoder[T]) (<-chan swarm.Msg[T], chan<- swarm.Msg[T]) {
	rcv := make(chan swarm.Msg[T], k.Config.CapRcv)
	ack := make(chan swarm.Msg[T], k.Config.CapAck)

	k.RWMutex.Lock()
	k.router[codec.Category()] = newMsgRouter[T](rcv, codec)
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
		slog.Debug("kernel dequeue started", "cat", codec.Category())
		defer slog.Debug("kernel dequeue stopped", "cat", codec.Category())

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

// RecvEvent creates pair of channels within kernel to receive events
func RecvEvent[E swarm.Event[M, T], M, T any](k *ListenerIO, codec Decoder[E]) (<-chan E, chan<- E) {
	rcv := make(chan E, k.Config.CapRcv)
	ack := make(chan E, k.Config.CapAck)

	k.RWMutex.Lock()
	k.router[codec.Category()] = newEvtRouter(rcv, codec)
	k.RWMutex.Unlock()

	shape := optics.ForShape2[E, swarm.Digest, error]()

	// emitter routine
	acks := func(evt E) {
		digest, exx := shape.Get(&evt)
		if exx == nil {
			err := k.Config.Backoff.Retry(
				func() error {
					return k.Listener.Ack(k.context, digest)
				},
			)
			if k.Config.StdErr != nil && err != nil {
				k.Config.StdErr <- swarm.ErrDequeue.With(err)
			}
		} else {
			err := k.Config.Backoff.Retry(
				func() error {
					return k.Listener.Err(k.context, digest, exx)
				},
			)
			if k.Config.StdErr != nil && err != nil {
				k.Config.StdErr <- swarm.ErrDequeue.With(err)
			}
		}
	}

	k.WaitGroup.Add(1)
	go func() {
		slog.Debug("kernel dequeue started", "cat", codec.Category())
		defer slog.Debug("kernel dequeue stopped", "cat", codec.Category())

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
			case evt := <-ack:
				acks(evt)
			}
		}

		backlog := len(ack)
		close(ack)

		if backlog != 0 {
			for evt := range ack {
				acks(evt)
			}
		}

		k.WaitGroup.Done()
	}()

	return rcv, ack
}
