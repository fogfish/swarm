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

	"github.com/fogfish/golem/optics"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/broadcast"
)

// Emitter defines on-the-wire protocol for [swarm.Bag], covering egress use-cases
type Emitter interface {
	Enq(context.Context, swarm.Bag) error
	Close() error
}

// Encodes message into wire format
type Encoder[T any] interface {
	Category() string
	Encode(T) (swarm.Bag, error)
}

// The egress part of the kernel is used to enqueue messages into message broker.
type EmitterCore struct {
	sync.WaitGroup

	// Control-plane stop channel used by go routines to stop I/O on data channels
	context context.Context
	cancel  context.CancelFunc

	// Control-plane to preempt emitter loop. It is used in externally scheduled
	// environment to guarantee that all emitted messages are sent to broker before application is preempted.
	ctrlPreempt *broadcast.Broadcaster

	// Kernel configuration
	Config swarm.Config

	// Emitter is the writer port on message broker
	Emitter Emitter
}

// Creates a new emitter kernel with the given emitter and configuration.
func NewEmitter(emitter Emitter, config swarm.Config) *EmitterCore {
	return builder().Emitter(emitter, config)
}

// Creates a new emitter kernel with the given emitter and configuration.
func newEmitter(emitter Emitter, config swarm.Config) *EmitterCore {
	ctx, can := context.WithCancel(context.Background())

	return &EmitterCore{
		Config:  config,
		context: ctx,
		cancel:  can,
		Emitter: emitter,
	}
}

// Close emitter
func (k *EmitterCore) Close() {
	k.cancel()
	k.WaitGroup.Wait()
	k.Emitter.Close()
}

// Await enqueue
func (k *EmitterCore) Await() {
	<-k.context.Done()
	k.WaitGroup.Wait()
	k.Emitter.Close()
}

// Creates pair of channels within kernel to emit messages to broker.
func EmitChan[T any](k *EmitterCore, codec Encoder[T]) (chan<- T, <-chan T) {
	snd := make(chan T, k.Config.CapOut)
	dlq := make(chan T, k.Config.CapDlq)

	var ctl chan chan struct{}
	if k.ctrlPreempt != nil {
		ctl = k.ctrlPreempt.Register()
	}

	// emitter routine
	emit := func(obj T) {
		bag, err := codec.Encode(obj)
		if err != nil {
			dlq <- obj
			if k.Config.StdErr != nil {
				k.Config.StdErr <- swarm.ErrEncoder.With(err)
			}
			slog.Debug("emitter failed to encode message",
				slog.Any("cat", codec.Category()),
				slog.Any("obj", obj),
				slog.Any("err", err),
			)
			return
		}

		err = k.Config.Backoff.Retry(func() error {
			return k.Emitter.Enq(context.Background(), bag)
		})
		if err != nil {
			dlq <- obj
			if k.Config.StdErr != nil {
				k.Config.StdErr <- swarm.ErrEnqueue.With(err)
			}
			slog.Debug("emitter failed to send message",
				slog.Any("cat", bag.Category),
				slog.Any("bag", bag),
				slog.Any("err", err),
			)
			return
		}
	}

	k.WaitGroup.Add(1)
	go func() {
		slog.Info("init message emitter", slog.Any("cat", codec.Category()))
		defer slog.Info("free message emitter", slog.Any("cat", codec.Category()))

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
			case sack := <-ctl:
				for range len(snd) {
					emit(<-snd)
				}
				func() {
					defer func() {
						recover() // Ignore panic if ackCh is closed due to timeout (preemption)
					}()
					sack <- struct{}{}
				}()
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

		if k.ctrlPreempt != nil {
			k.ctrlPreempt.Unregister(ctl)
		}
		k.WaitGroup.Done()
	}()

	return snd, dlq
}

// Creates pair of channels within kernel to events to broker.
func EmitEvent[E swarm.Event[M, T], M, T any](k *EmitterCore, codec Encoder[E]) (chan<- E, <-chan E) {
	snd := make(chan E, k.Config.CapOut)
	dlq := make(chan E, k.Config.CapDlq)

	var ctl chan chan struct{}
	if k.ctrlPreempt != nil {
		ctl = k.ctrlPreempt.Register()
	}

	shape := optics.ForProduct1[E, error]()

	// emitter routine
	emit := func(evt E) {
		bag, err := codec.Encode(evt)
		if err != nil {
			shape.Put(&evt, err)
			dlq <- evt
			if k.Config.StdErr != nil {
				k.Config.StdErr <- swarm.ErrEncoder.With(err)
			}
			slog.Debug("emitter failed to encode event",
				slog.Any("cat", codec.Category()),
				slog.Any("evt", evt),
				slog.Any("err", err),
			)
			return
		}

		err = k.Config.Backoff.Retry(func() error {
			return k.Emitter.Enq(context.Background(), bag)
		})
		if err != nil {
			shape.Put(&evt, err)
			dlq <- evt
			if k.Config.StdErr != nil {
				k.Config.StdErr <- swarm.ErrEnqueue.With(err)
			}
			slog.Debug("emitter failed to send event",
				slog.Any("cat", bag.Category),
				slog.Any("bag", bag),
				slog.Any("err", err),
			)
			return
		}
	}

	k.WaitGroup.Add(1)
	go func() {
		slog.Info("init event emitter", slog.Any("cat", codec.Category()))
		defer slog.Info("free event emitter", slog.Any("cat", codec.Category()))

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
			case sack := <-ctl:
				for range len(snd) {
					emit(<-snd)
				}
				func() {
					defer func() {
						recover() // Ignore panic if ackCh is closed due to timeout (preemption)
					}()
					sack <- struct{}{}
				}()
			case evt := <-snd:
				emit(evt)
			}
		}

		backlog := len(snd)
		close(snd)

		if backlog != 0 {
			for evt := range snd {
				emit(evt)
			}
		}

		if k.ctrlPreempt != nil {
			k.ctrlPreempt.Unregister(ctl)
		}
		k.WaitGroup.Done()
	}()

	return snd, dlq
}
