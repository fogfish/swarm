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
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
)

func TestDequeuer(t *testing.T) {
	codec := encoding.NewCodecJson[string]()
	none := mockCathode(nil, nil)
	pass := mockCathode(make(chan string),
		[]swarm.Bag{{Ctx: &swarm.Context{Category: "test", Digest: "1"}, Object: []byte(`"1"`)}},
	)

	t.Run("Kernel", func(t *testing.T) {
		k := New(nil, NewDequeuer(mockCathode(nil, nil), swarm.Config{}))
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

//------------------------------------------------------------------------------

type cathode struct {
	seq []swarm.Bag
	ack chan string
}

func mockCathode(ack chan string, seq []swarm.Bag) cathode {
	return cathode{seq: seq, ack: ack}
}

func (c cathode) Ack(ctx context.Context, digest string) error {
	c.ack <- digest
	return nil
}

func (c cathode) Err(ctx context.Context, digest string, err error) error {
	c.ack <- digest
	return nil
}

func (c cathode) Ask(context.Context) ([]swarm.Bag, error) {
	return c.seq, nil
}
