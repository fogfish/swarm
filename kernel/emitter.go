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

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/broadcast"
)

// Emitter defines on-the-wire protocol for [swarm.Bag], covering egress use-cases
type Emitter interface {
	Enq(context.Context, swarm.Bag) error
}

// Encodes message into wire format
type Encoder[T any] interface {
	Category() string
	Encode(T) ([]byte, error)
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
}

// Await enqueue
func (k *EmitterCore) Await() {
	<-k.context.Done()
	k.WaitGroup.Wait()
}

// Creates pair of channels within kernel to emit messages to broker.
func EmitChan[T any](k *EmitterCore, cat string, codec Encoder[T]) ( /*snd*/ chan<- T /*dlq*/, <-chan T) {
	snd := make(chan T, k.Config.CapOut)
	dlq := make(chan T, k.Config.CapDlq)

	var ctl chan chan struct{}
	if k.ctrlPreempt != nil {
		ctl = k.ctrlPreempt.Register()
	}

	// emitter routine
	emit := func(obj T) {
		msg, err := codec.Encode(obj)
		if err != nil {
			dlq <- obj
			if k.Config.StdErr != nil {
				k.Config.StdErr <- swarm.ErrEncoder.With(err)
			}
			slog.Debug("emitter failed to encode message",
				slog.Any("cat", cat),
				slog.Any("obj", obj),
				slog.Any("err", err),
			)
			return
		}

		bag := swarm.Bag{Category: cat, Object: msg}
		err = k.Config.Backoff.Retry(func() error {
			return k.Emitter.Enq(context.Background(), bag)
		})
		if err != nil {
			dlq <- obj
			if k.Config.StdErr != nil {
				k.Config.StdErr <- swarm.ErrEnqueue.With(err)
			}
			slog.Debug("emitter failed to send message",
				slog.Any("cat", cat),
				slog.Any("bag", bag),
				slog.Any("err", err),
			)
			return
		}
	}

	k.WaitGroup.Add(1)
	go func() {
		slog.Info("init emitter", slog.Any("cat", cat))
		defer slog.Info("free emitter", slog.Any("cat", cat))

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
