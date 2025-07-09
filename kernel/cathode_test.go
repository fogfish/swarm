//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import (
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
)

func TestDequeuer(t *testing.T) {
	codec := encoding.ForTyped[string]()
	mock := mockFactory{}
	none := mock.Cathode(nil, nil)
	pass := mock.Cathode(make(chan string), mock.Bag(1))

	t.Run("Kernel", func(t *testing.T) {
		k := New(nil, NewDequeuer(newMockCathode(nil, nil), swarm.Config{}))
		go func() {
			time.Sleep(yield_before_close)
			k.Close()
		}()
		k.Await()
	})

	t.Run("None", func(t *testing.T) {
		k := NewDequeuer(none, swarm.Config{PollFrequency: 1 * time.Second})
		go k.Await()
		k.Close()
	})

	t.Run("Idle", func(t *testing.T) {
		k := NewDequeuer(none, swarm.Config{PollFrequency: 1 * time.Second})
		Dequeue(k, "test", codec)
		go k.Await()
		k.Close()
	})

	t.Run("Dequeue.1", func(t *testing.T) {
		k := NewDequeuer(pass, swarm.Config{PollFrequency: 10 * time.Millisecond})
		rcv, ack := Dequeue(k, "test", codec)
		go k.Await()

		ack <- <-rcv
		it.Then(t).Should(
			it.Equal(string(<-pass.ack), `1`),
		)

		k.Close()
	})

	t.Run("Dequeue.1.Context", func(t *testing.T) {
		k := NewDequeuer(pass, swarm.Config{PollFrequency: 10 * time.Millisecond})
		rcv, ack := Dequeue(k, "test", codec)
		go k.Await()

		msg := <-rcv
		ack <- msg
		it.Then(t).Should(
			it.Equal(string(<-pass.ack), `1`),
			it.Equal(msg.IOContext.(string), "context"),
		)

		k.Close()
	})

	t.Run("Backlog", func(t *testing.T) {
		k := NewDequeuer(pass, swarm.Config{CapAck: 4, PollFrequency: 1 * time.Millisecond})
		rcv, ack := Dequeue(k, "test", codec)
		go k.Await()

		ack <- <-rcv
		ack <- <-rcv
		ack <- <-rcv
		ack <- <-rcv
		go k.Close()

		it.Then(t).Should(
			it.Equal(string(<-pass.ack), `1`),
			it.Equal(string(<-pass.ack), `1`),
			it.Equal(string(<-pass.ack), `1`),
			it.Equal(string(<-pass.ack), `1`),
		)
	})
}
