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
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/fogfish/swarm"
)

type Codec[T any] interface {
	Encode(T) ([]byte, error)
	Decode([]byte) (T, error)
}

type Emitter interface {
	Enq(swarm.Bag) error
}

type Cathode interface {
	Ack(digest string) error
	Ask() ([]swarm.Bag, error)
}

type Spawner interface {
	Spawn(*Kernel) error
}

// Event Kernel builds an infrastructure for integrating message brokers,
// events busses into Golang channel environment.
// The implementation follows the pattern, defined by
// https://go101.org/article/channel-closing.html
type Kernel struct {
	sync.WaitGroup
	sync.RWMutex

	// Kernel configuartion
	Config swarm.Config

	// Control-plane stop channel. It is used to notify the kernel to terminate.
	// Kernel notifies control plane of individual routines.
	mainStop chan struct{}

	// Control-plane stop channel used by go routines to stop I/O on data channels
	ctrlStop chan struct{}

	// Control-plane for ack of batch elements
	ctrlAcks chan *swarm.Context

	// event router, binds category with destination channel
	router map[string]interface{ Route(swarm.Bag) error }

	// The flag indicates if Await loop is started
	waiting bool

	// On the wire protocol emitter (writer) and cathode (receiver)
	Emitter Emitter
	Cathode Cathode
}

// New routing and dispatch kernel
func New(emitter Emitter, cathode Cathode, config swarm.Config) *Kernel {
	return &Kernel{
		Config:   config,
		mainStop: make(chan struct{}, 1), // MUST BE buffered
		ctrlStop: make(chan struct{}),

		router: map[string]interface{ Route(swarm.Bag) error }{},

		Emitter: emitter,
		Cathode: cathode,
	}
}

// internal infinite receive loop.
// waiting for message from event buses and queues and schedules it for delivery.
func (k *Kernel) receive() {
	k.WaitGroup.Add(1)

	asker := func() {
		seq, err := k.Cathode.Ask()
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
				err := r.Route(bag)
				if k.Config.StdErr != nil && err != nil {
					k.Config.StdErr <- err
					return
				}
			}
		}
	}

	go func() {
	exit:
		for {
			select {
			case <-k.ctrlStop:
				break exit
			default:
			}

			select {
			case <-k.ctrlStop:
				break exit
			case <-time.After(k.Config.PollFrequency):
				asker()
			}
		}

		slog.Debug("Free kernel infinite loop")
		k.WaitGroup.Done()
	}()

	slog.Debug("Init kernel infinite loop")
}

// Close event delivery infrastructure
func (k *Kernel) Close() {
	k.mainStop <- struct{}{}
	if !k.waiting {
		close(k.ctrlStop)
		k.WaitGroup.Wait()
	}
}

// Await for event delivery
func (k *Kernel) Await() {
	k.waiting = true

	if spawner, ok := k.Cathode.(Spawner); ok {
		spawner.Spawn(k)
	} else {
		k.receive()
	}

	<-k.mainStop
	close(k.ctrlStop)
	k.WaitGroup.Wait()
}

// Dispatches batch of messages
func (k *Kernel) Dispatch(seq []swarm.Bag, timeout time.Duration) error {
	k.WaitGroup.Add(1)
	k.ctrlAcks = make(chan *swarm.Context, len(seq))

	wnd := map[string]struct{}{}
	for i := 0; i < len(seq); i++ {
		bag := seq[i]
		wnd[bag.Ctx.Digest] = struct{}{}

		k.RWMutex.RLock()
		r, has := k.router[bag.Ctx.Category]
		k.RWMutex.RUnlock()

		if has {
			err := r.Route(bag)
			if k.Config.StdErr != nil && err != nil {
				k.Config.StdErr <- err
				continue
			}
		}
	}

	var err error

exit:
	for {
		select {
		case <-k.ctrlStop:
			break exit
		default:
		}

		select {
		case <-k.ctrlStop:
			break exit
		case ack := <-k.ctrlAcks:
			if err == nil && ack.Error != nil {
				err = ack.Error
			}

			delete(wnd, ack.Digest)
			if len(wnd) == 0 {
				break exit
			}
		case <-time.After(timeout):
			err = errors.New("ack timeout")
			break exit
		}
	}

	close(k.ctrlAcks)
	k.ctrlAcks = nil

	return err
}

// Enqueue channels for kernel
func Enqueue[T any](k *Kernel, cat string, codec Codec[T]) ( /*snd*/ chan<- T /*dlq*/, <-chan T) {
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

		if err := k.Emitter.Enq(bag); err != nil {
			dlq <- obj
			if k.Config.StdErr != nil {
				k.Config.StdErr <- err
			}
			return
		}

		slog.Debug("emitted ", "cat", cat, "object", obj)
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
			case <-k.ctrlStop:
				break exit
			default:
			}

			select {
			case <-k.ctrlStop:
				break exit
			case obj := <-snd:
				emit(obj)
			}
		}

		backlog := len(snd)
		close(snd)

		slog.Debug("Free enqueue kernel", "cat", cat, "backlog", backlog)
		if backlog != 0 {
			for obj := range snd {
				emit(obj)
			}
		}
		k.WaitGroup.Done()
	}()

	slog.Debug("Init enqueue kernel", "cat", cat)

	return snd, dlq
}

type router[T any] struct {
	ch    chan swarm.Msg[T]
	codec Codec[T]
}

func (a router[T]) Route(bag swarm.Bag) error {
	obj, err := a.codec.Decode(bag.Object)
	if err != nil {
		return err
	}

	msg := swarm.Msg[T]{Ctx: bag.Ctx, Object: obj}
	a.ch <- msg
	return nil
}

// Enqueue channels for kernel
func Dequeue[T any](k *Kernel, cat string, codec Codec[T]) ( /*rcv*/ <-chan swarm.Msg[T] /*ack*/, chan<- swarm.Msg[T]) {
	rcv := make(chan swarm.Msg[T], k.Config.CapRcv)
	ack := make(chan swarm.Msg[T], k.Config.CapAck)

	k.RWMutex.Lock()
	k.router[cat] = router[T]{ch: rcv, codec: codec}
	k.RWMutex.Unlock()

	// emitter routine
	acks := func(msg swarm.Msg[T]) {
		if msg.Ctx.Error == nil {
			err := k.Cathode.Ack(msg.Ctx.Digest)
			if k.Config.StdErr != nil && err != nil {
				k.Config.StdErr <- err
			}

			slog.Debug("acked ", "cat", cat, "object", msg.Object)
		}

		if k.ctrlAcks != nil {
			k.ctrlAcks <- msg.Ctx
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
			case <-k.ctrlStop:
				break exit
			default:
			}

			select {
			case <-k.ctrlStop:
				break exit
			case msg := <-ack:
				acks(msg)
			}
		}

		backlog := len(ack)
		close(ack)

		slog.Debug("Free dequeue kernel", "cat", cat, "backlog", backlog)
		if backlog != 0 {
			for msg := range ack {
				acks(msg)
			}
		}
		k.WaitGroup.Done()
	}()

	slog.Debug("Init dequeue kernel", "cat", cat)

	return rcv, ack
}
