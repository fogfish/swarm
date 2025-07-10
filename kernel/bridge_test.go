//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import (
	"fmt"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm/kernel/encoding"
)

// controls yield time before kernel is closed
const yield_before_close = 50 * time.Millisecond

func TestBridge(t *testing.T) {
	cfg := newConfig()
	// Config Kernel for extreme thresholds
	cfg.kernel.PollFrequency = 10 * time.Nanosecond
	cfg.kernel.TimeToFlight = 2 * time.Millisecond

	codec := encoding.ForTyped[string]()
	mock := mockFactory{}

	t.Run("None", func(t *testing.T) {
		k := mock.Dequeuer(mock.Bridge(cfg, mock.Bag(1)), cfg)
		Dequeue(k, "test", codec)
		k.Await()
	})

	t.Run("Dequeue.1", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(1))
		k := mock.Dequeuer(bridge, cfg)

		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() { ack <- <-rcv }()
		k.Await()

		it.Then(t).Should(
			it.Seq(bridge.ack).Equal(`1`),
		)
	})

	t.Run("Dequeue.N", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(3))
		k := mock.Dequeuer(bridge, cfg)

		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			ack <- <-rcv
			ack <- <-rcv
		}()
		k.Await()

		it.Then(t).Should(
			it.Seq(bridge.ack).Equal(`1`, `2`, `3`),
		)
	})

	t.Run("Error.1", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(1))
		k := mock.Dequeuer(bridge, cfg)

		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			x := <-rcv
			ack <- x.Fail(fmt.Errorf("failed"))
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(bridge.Status).Contain("failed"),
		)
	})

	t.Run("Error.N.1", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(3))
		k := mock.Dequeuer(bridge, cfg)

		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			x := <-rcv
			ack <- x.Fail(fmt.Errorf("failed"))
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(bridge.Status).Contain("failed"),
		)
	})

	t.Run("Error.N.2", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(3))
		k := mock.Dequeuer(bridge, cfg)

		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			x := <-rcv
			ack <- x.Fail(fmt.Errorf("failed"))
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(bridge.Status).Contain("failed"),
		)
	})

	t.Run("Error.N.3", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(3))
		k := mock.Dequeuer(bridge, cfg)

		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			ack <- <-rcv
			x := <-rcv
			ack <- x.Fail(fmt.Errorf("failed"))
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(bridge.Status).Contain("failed"),
		)
	})

	t.Run("Timeout.1", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(1))
		k := mock.Dequeuer(bridge, cfg)

		rcv, _ := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() { <-rcv }()
		k.Await()

		it.Then(t).Should(
			it.Fail(bridge.Status).Contain("timeout"),
		)
	})

	t.Run("Timeout.N.1", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(3))
		k := mock.Dequeuer(bridge, cfg)

		rcv, _ := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			<-rcv
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(bridge.Status).Contain("timeout"),
		)
	})

	t.Run("Timeout.N.2", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(3))
		k := mock.Dequeuer(bridge, cfg)

		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			<-rcv
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(bridge.Status).Contain("timeout"),
		)
	})

	t.Run("Timeout.N.3", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(3))
		k := mock.Dequeuer(bridge, cfg)

		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			ack <- <-rcv
			<-rcv
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(bridge.Status).Contain("timeout"),
		)
	})

}

func TestBridgeWait(t *testing.T) {
	cfg := newConfig()
	cfg.kernel.PollFrequency = 10 * time.Nanosecond
	cfg.kernel.TimeToFlight = 10 * time.Millisecond // making extreme for testing

	codec := encoding.ForTyped[string]()
	mock := mockFactory{}
	mock.enableExternalPreemption()

	t.Run("Enqueue.After.Dequeue.1", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(1))
		deq := mock.Dequeuer(bridge, cfg)

		emit := mock.Emitter(cfg)
		enq := mock.Enqueuer(emit, cfg)

		rcv, ack := Dequeue(deq, "test", codec)
		snd, _ := Enqueue(enq, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			x := <-rcv
			snd <- x.Object
			ack <- x
		}()
		deq.Await()

		it.Then(t).Should(
			it.Nil(bridge.Status()),
			it.Seq(bridge.ack).Equal(`1`),
			it.Seq(emit.seq).Equal(`"1"`),
			it.Equal(emit.emittedAt.Compare(bridge.stoppedAt), -1),
		)
	})

	t.Run("Enqueue.After.Dequeue.N", func(t *testing.T) {
		bridge := mock.Bridge(cfg, mock.Bag(1))
		deq := mock.Dequeuer(bridge, cfg)

		emit := mock.Emitter(cfg)
		enq := mock.Enqueuer(emit, cfg)

		rcv, ack := Dequeue(deq, "test", codec)
		snd, _ := Enqueue(enq, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			x := <-rcv
			snd <- x.Object
			snd <- x.Object
			snd <- x.Object
			snd <- x.Object
			ack <- x
		}()
		deq.Await()

		it.Then(t).Should(
			it.Nil(bridge.Status()),
			it.Seq(bridge.ack).Equal(`1`),
			it.Seq(emit.seq).Equal(`"1"`, `"1"`, `"1"`, `"1"`),
			it.Equal(emit.emittedAt.Compare(bridge.stoppedAt), -1),
		)
	})
}

func TestBridgeWaitTimeout(t *testing.T) {
	cfg := newConfig()
	cfg.timeToEmit = 10 * time.Millisecond
	cfg.kernel.PollFrequency = 10 * time.Nanosecond
	cfg.kernel.TimeToFlight = 2 * time.Millisecond

	codec := encoding.ForTyped[string]()
	mock := mockFactory{}
	mock.enableExternalPreemption()

	bridge := mock.Bridge(cfg, mock.Bag(1))
	deq := mock.Dequeuer(bridge, cfg)

	emit := mock.Emitter(cfg)
	enq := mock.Enqueuer(emit, cfg)

	rcv, ack := Dequeue(deq, "test", codec)
	snd, _ := Enqueue(enq, "test", codec)

	// Note: in real apps receive loop is always go function
	go func() {
		x := <-rcv
		snd <- x.Object
		ack <- x
	}()
	deq.Await()

	it.Then(t).Should(
		it.Fail(bridge.Status).Contain("deadline"),
	)
}
