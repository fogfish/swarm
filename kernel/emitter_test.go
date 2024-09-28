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
	"fmt"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
)

func TestEnqueuer(t *testing.T) {
	codec := encoding.NewCodecJson[string]()
	mockit := func(config swarm.Config) (*Enqueuer, *emitter) {
		mock := mockEmitter(10)
		k := NewEnqueuer(mock, config)

		go func() {
			time.Sleep(yield_before_close)
			k.Close()
		}()

		return k, mock
	}

	t.Run("Kernel", func(t *testing.T) {
		k := New(NewEnqueuer(mockEmitter(10), swarm.Config{}), nil)
		go func() {
			time.Sleep(yield_before_close)
			k.Close()
		}()
		k.Await()
	})

	t.Run("None", func(t *testing.T) {
		k, _ := mockit(swarm.Config{})
		Enqueue(k, "test", codec)
		k.Await()
	})

	t.Run("Enqueue.1", func(t *testing.T) {
		k, e := mockit(swarm.Config{})
		snd, _ := Enqueue(k, "test", codec)

		snd <- "1"
		it.Then(t).Should(
			it.Equal(<-e.val, `"1"`),
		)

		k.Await()
	})

	t.Run("Enqueue.1.Shut", func(t *testing.T) {
		k, e := mockit(swarm.Config{})
		snd, _ := Enqueue(k, "test", codec)

		snd <- "1"
		k.Await()

		it.Then(t).Should(
			it.Seq(e.seq).Equal(`"1"`),
		)
	})

	t.Run("Enqueue.1.Error", func(t *testing.T) {
		err := make(chan error)
		k := NewEnqueuer(looser{}, swarm.Config{StdErr: err})
		snd, dlq := Enqueue(k, "test", codec)

		snd <- "1"
		it.Then(t).Should(
			it.Equal(<-dlq, "1"),
			it.Fail(func() error { return <-err }).Contain("lost"),
		)

		k.Close()
	})

	t.Run("Enqueue.1.Codec", func(t *testing.T) {
		err := make(chan error)
		k := NewEnqueuer(mockEmitter(10), swarm.Config{StdErr: err})
		snd, dlq := Enqueue(k, "test", looser{})

		snd <- "1"
		it.Then(t).Should(
			it.Equal(<-dlq, "1"),
			it.Fail(func() error { return <-err }).Contain("invalid"),
		)

		k.Close()
	})

	t.Run("Enqueue.N", func(t *testing.T) {
		k, e := mockit(swarm.Config{})
		snd, _ := Enqueue(k, "test", codec)

		snd <- "1"
		it.Then(t).Should(
			it.Equal(<-e.val, `"1"`),
		)

		snd <- "2"
		it.Then(t).Should(
			it.Equal(<-e.val, `"2"`),
		)

		snd <- "3"
		it.Then(t).Should(
			it.Equal(<-e.val, `"3"`),
		)

		k.Await()
	})

	t.Run("Enqueue.N.Shut", func(t *testing.T) {
		k, e := mockit(swarm.Config{})
		snd, _ := Enqueue(k, "test", codec)

		snd <- "1"
		snd <- "2"
		snd <- "3"

		k.Await()

		it.Then(t).Should(
			it.Seq(e.seq).Equal(`"1"`, `"2"`, `"3"`),
		)
	})

	t.Run("Enqueue.N.Backlog", func(t *testing.T) {
		e := mockEmitter(10)
		k := NewEnqueuer(e, swarm.Config{CapOut: 4})
		snd, _ := Enqueue(k, "test", codec)

		snd <- "1"
		snd <- "2"
		snd <- "3"

		k.Close()

		it.Then(t).Should(
			it.Seq(e.seq).Equal(`"1"`, `"2"`, `"3"`),
		)
	})

	t.Run("Enqueue.N.Error", func(t *testing.T) {
		err := make(chan error)
		k := NewEnqueuer(looser{}, swarm.Config{CapOut: 4, CapDLQ: 4, StdErr: err})
		snd, dlq := Enqueue(k, "test", codec)

		snd <- "1"
		snd <- "2"
		snd <- "3"

		it.Then(t).Should(
			it.Equal(<-dlq, "1"),
			it.Fail(func() error { return <-err }).Contain("lost"),

			it.Equal(<-dlq, "2"),
			it.Fail(func() error { return <-err }).Contain("lost"),

			it.Equal(<-dlq, "3"),
			it.Fail(func() error { return <-err }).Contain("lost"),
		)

		k.Close()
	})
}

//------------------------------------------------------------------------------

type emitter struct {
	wait int
	seq  []string
	val  chan string
}

func mockEmitter(wait int) *emitter {
	return &emitter{
		wait: wait,
		seq:  make([]string, 0),
		val:  make(chan string, 1000),
	}
}

func (e *emitter) Enq(ctx context.Context, bag swarm.Bag) error {
	time.Sleep(time.Duration(e.wait) * time.Microsecond)
	e.seq = append(e.seq, string(bag.Object))

	e.val <- string(bag.Object)
	return nil
}

type looser struct{}

func (e looser) Enq(ctx context.Context, bag swarm.Bag) error {
	return fmt.Errorf("lost")
}

func (e looser) Encode(x string) ([]byte, error) {
	return nil, fmt.Errorf("invalid")
}
