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
)

func TestEnqueuer(t *testing.T) {
	none := mockEmitter(0, nil)
	pass := mockEmitter(10, make(chan []byte))

	t.Run("None", func(t *testing.T) {
		k := NewEnqueuer(none, swarm.Config{})
		k.Close()
	})

	t.Run("Idle", func(t *testing.T) {
		k := NewEnqueuer(none, swarm.Config{})
		Enqueue(k, "test", swarm.NewCodecJson[string]())
		k.Close()
	})

	t.Run("Enqueue.1", func(t *testing.T) {
		k := NewEnqueuer(pass, swarm.Config{})
		snd, _ := Enqueue(k, "test", swarm.NewCodecJson[string]())

		snd <- "1"
		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"1"`),
		)

		k.Close()
	})

	t.Run("Enqueue.N", func(t *testing.T) {
		k := NewEnqueuer(pass, swarm.Config{})
		snd, _ := Enqueue(k, "test", swarm.NewCodecJson[string]())

		snd <- "1"
		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"1"`),
		)

		snd <- "2"
		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"2"`),
		)

		snd <- "3"
		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"3"`),
		)

		snd <- "4"
		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"4"`),
		)

		k.Close()
	})

	t.Run("Backlog", func(t *testing.T) {
		k := NewEnqueuer(pass, swarm.Config{CapOut: 4})
		snd, _ := Enqueue(k, "test", swarm.NewCodecJson[string]())
		snd <- "1"
		snd <- "2"
		snd <- "3"
		snd <- "4"
		go k.Close()

		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"1"`),
			it.Equal(string(<-pass.ch), `"2"`),
			it.Equal(string(<-pass.ch), `"3"`),
			it.Equal(string(<-pass.ch), `"4"`),
		)
	})

	t.Run("Queues.N", func(t *testing.T) {
		k := NewEnqueuer(pass, swarm.Config{})
		a, _ := Enqueue(k, "a", swarm.NewCodecJson[string]())
		b, _ := Enqueue(k, "b", swarm.NewCodecJson[string]())
		c, _ := Enqueue(k, "c", swarm.NewCodecJson[string]())

		a <- "a"
		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"a"`),
		)

		b <- "b"
		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"b"`),
		)

		c <- "c"
		it.Then(t).Should(
			it.Equal(string(<-pass.ch), `"c"`),
		)

		k.Close()
	})
}

//------------------------------------------------------------------------------

type emitter struct {
	ms int
	ch chan []byte
}

func mockEmitter(ms int, ch chan []byte) emitter {
	return emitter{
		ms: ms,
		ch: ch,
	}
}

func (e emitter) Enq(ctx context.Context, bag swarm.Bag) error {
	time.Sleep(time.Duration(e.ms) * time.Microsecond)
	e.ch <- bag.Object
	return nil
}
