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
	cfg := newConfig()

	codec := encoding.ForTyped[string]()
	mock := mockFactory{}

	t.Run("Kernel", func(t *testing.T) {
		emit := mock.Emitter(cfg)
		k := New(NewEmitter(emit, cfg.kernel), nil)

		go func() {
			time.Sleep(cfg.yieldBeforeClose)
			k.Close()
		}()
		k.Await()
	})

	t.Run("None", func(t *testing.T) {
		emit := mock.Emitter(cfg)
		k := mock.Enqueuer(emit, cfg)

		EmitChan(k, "test", codec)
		k.Await()
	})

	t.Run("Enqueue.1", func(t *testing.T) {
		emit := mock.Emitter(cfg)
		k := mock.Enqueuer(emit, cfg)

		snd, _ := EmitChan(k, "test", codec)

		snd <- "1"
		it.Then(t).Should(
			it.Equal(<-emit.val, `"1"`),
		)

		k.Await()
	})

	t.Run("Enqueue.1.Shut", func(t *testing.T) {
		emit := mock.Emitter(cfg)
		k := mock.Enqueuer(emit, cfg)

		snd, _ := EmitChan(k, "test", codec)

		snd <- "1"
		k.Await()

		it.Then(t).Should(
			it.Seq(emit.seq).Equal(`"1"`),
		)
	})

	t.Run("Enqueue.1.Error", func(t *testing.T) {
		cfg := newConfig()
		err := make(chan error)
		cfg.kernel.StdErr = err
		k := mock.Enqueuer(looser{}, cfg)

		snd, dlq := EmitChan(k, "test", codec)

		snd <- "1"
		it.Then(t).Should(
			it.Equal(<-dlq, "1"),
			it.Fail(func() error { return <-err }).Contain("lost"),
		)

		k.Close()
	})

	t.Run("Enqueue.1.Codec", func(t *testing.T) {
		cfg := newConfig()
		err := make(chan error)
		cfg.kernel.StdErr = err
		k := mock.Enqueuer(looser{}, cfg)

		snd, dlq := EmitChan(k, "test", looser{})

		snd <- "1"
		it.Then(t).Should(
			it.Equal(<-dlq, "1"),
			it.Fail(func() error { return <-err }).Contain("invalid"),
		)

		k.Close()
	})

	t.Run("Enqueue.N", func(t *testing.T) {
		emit := mock.Emitter(cfg)
		k := mock.Enqueuer(emit, cfg)

		snd, _ := EmitChan(k, "test", codec)

		snd <- "1"
		it.Then(t).Should(
			it.Equal(<-emit.val, `"1"`),
		)

		snd <- "2"
		it.Then(t).Should(
			it.Equal(<-emit.val, `"2"`),
		)

		snd <- "3"
		it.Then(t).Should(
			it.Equal(<-emit.val, `"3"`),
		)

		k.Await()
	})

	t.Run("Enqueue.N.Shut", func(t *testing.T) {
		emit := mock.Emitter(cfg)
		k := mock.Enqueuer(emit, cfg)

		snd, _ := EmitChan(k, "test", codec)

		snd <- "1"
		snd <- "2"
		snd <- "3"

		k.Await()

		it.Then(t).Should(
			it.Seq(emit.seq).Equal(`"1"`, `"2"`, `"3"`),
		)
	})

	t.Run("Enqueue.N.Backlog", func(t *testing.T) {
		cfg := newConfig()
		cfg.kernel.CapOut = 4

		emit := mock.Emitter(cfg)
		k := mock.Enqueuer(emit, cfg)

		snd, _ := EmitChan(k, "test", codec)

		snd <- "1"
		snd <- "2"
		snd <- "3"

		k.Close()

		it.Then(t).Should(
			it.Seq(emit.seq).Equal(`"1"`, `"2"`, `"3"`),
		)
	})

	t.Run("Enqueue.N.Error", func(t *testing.T) {
		err := make(chan error)
		cfg := newConfig()
		cfg.kernel.CapOut = 4
		cfg.kernel.CapDlq = 4
		cfg.kernel.StdErr = err

		k := mock.Enqueuer(looser{}, cfg)

		snd, dlq := EmitChan(k, "test", codec)

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

type looser struct{}

func (e looser) Enq(ctx context.Context, bag swarm.Bag) error {
	return fmt.Errorf("lost")
}

func (e looser) Category() string { return "test" }

func (e looser) Encode(x string) ([]byte, error) {
	return nil, fmt.Errorf("invalid")
}
