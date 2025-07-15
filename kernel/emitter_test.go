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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
)

func TestEmitterFactory(t *testing.T) {
	cfg := newConfig()
	mock := mockFactory{}

	t.Run("Kernel", func(t *testing.T) {
		emit := mock.EmitterCore(cfg)
		k := New(NewEmitter(emit, cfg.kernel), nil)

		go func() {
			time.Sleep(cfg.yieldBeforeClose)
			k.Close()
		}()
		k.Await()
	})
}

func TestEmitChan(t *testing.T) {
	emitTest(t,
		encoding.ForTyped[string](),
		EmitChan,
		func(i int) string {
			return fmt.Sprintf(`"%d"`, i)
		},
	)
}

func TestEmitEvent(t *testing.T) {
	type E = swarm.Event[swarm.Meta, string]

	emitTest(t,
		encoding.ForEvent[E]("testReal", "testAgent"),
		EmitEvent,
		func(i int) E {
			return E{
				Meta: &swarm.Meta{Type: "string"},
				Data: aws.String(fmt.Sprintf(`"%d"`, i)),
			}
		},
	)
}

func emitTest[T any](
	t *testing.T,
	codec Encoder[T],
	emitChan func(k *EmitterCore, codec Encoder[T]) (chan<- T, <-chan T),
	gen func(int) T,
) {
	t.Helper()
	conf := newConfig()
	mock := mockFactory{}

	t.Run("None", func(t *testing.T) {
		emit := mock.EmitterCore(conf)
		k := mock.Emitter(emit, conf)

		emitChan(k, codec)
		k.Await()
	})

	t.Run("Enqueue.1", func(t *testing.T) {
		emit := mock.EmitterCore(conf)
		k := mock.Emitter(emit, conf)

		snd, _ := emitChan(k, codec)

		obj := gen(1)
		snd <- obj
		it.Then(t).Should(
			it.Json(obj).Equiv(<-emit.val),
		)

		k.Await()
	})

	t.Run("Enqueue.1.Shutdown", func(t *testing.T) {
		emit := mock.EmitterCore(conf)
		k := mock.Emitter(emit, conf)

		snd, _ := emitChan(k, codec)

		obj := gen(1)
		snd <- obj
		k.Await()

		it.Then(t).Should(
			it.Equal(len(emit.seq), 1),
			it.Json(obj).Equiv(emit.seq[0]),
		)
	})

	t.Run("Enqueue.1.Error", func(t *testing.T) {
		cfg := newConfig()
		err := make(chan error)
		cfg.kernel.StdErr = err
		k := mock.Emitter(devnil[T]{}, cfg)

		snd, dlq := emitChan(k, codec)

		obj := gen(1)
		snd <- obj
		it.Then(t).Should(
			it.Equiv(obj, <-dlq),
			it.Fail(func() error { return <-err }).Contain("lost"),
		)

		k.Close()
	})

	t.Run("Enqueue.1.Codec", func(t *testing.T) {
		cfg := newConfig()
		err := make(chan error)
		cfg.kernel.StdErr = err
		k := mock.Emitter(devnil[T]{}, cfg)

		snd, dlq := emitChan(k, devnil[T]{})

		obj := gen(1)
		snd <- obj
		it.Then(t).Should(
			it.Equiv(obj, <-dlq),
			it.Fail(func() error { return <-err }).Contain("invalid"),
		)

		k.Close()
	})

	t.Run("Enqueue.N", func(t *testing.T) {
		emit := mock.EmitterCore(conf)
		k := mock.Emitter(emit, conf)

		snd, _ := emitChan(k, codec)

		obj := gen(1)
		snd <- obj
		it.Then(t).Should(
			it.Json(obj).Equiv(<-emit.val),
		)

		obj = gen(2)
		snd <- obj
		it.Then(t).Should(
			it.Json(obj).Equiv(<-emit.val),
		)

		obj = gen(3)
		snd <- obj
		it.Then(t).Should(
			it.Json(obj).Equiv(<-emit.val),
		)

		k.Await()
	})

	t.Run("Enqueue.N.Shutdown", func(t *testing.T) {
		emit := mock.EmitterCore(conf)
		k := mock.Emitter(emit, conf)

		snd, _ := emitChan(k, codec)

		seq := []T{gen(1), gen(2), gen(3)}
		for _, obj := range seq {
			snd <- obj
		}
		k.Await()

		for i, obj := range seq {
			it.Then(t).Should(
				it.Json(obj).Equiv(emit.seq[i]),
			)
		}
	})

	t.Run("Enqueue.N.Backlog", func(t *testing.T) {
		cfg := newConfig()
		cfg.kernel.CapOut = 4

		emit := mock.EmitterCore(cfg)
		k := mock.Emitter(emit, cfg)

		snd, _ := emitChan(k, codec)

		seq := []T{gen(1), gen(2), gen(3)}
		for _, obj := range seq {
			snd <- obj
		}
		k.Close()

		for i, obj := range seq {
			it.Then(t).Should(
				it.Json(obj).Equiv(emit.seq[i]),
			)
		}
	})

	t.Run("Enqueue.N.Error", func(t *testing.T) {
		err := make(chan error)
		cfg := newConfig()
		cfg.kernel.CapOut = 4
		cfg.kernel.CapDlq = 4
		cfg.kernel.StdErr = err

		k := mock.Emitter(devnil[T]{}, cfg)

		snd, dlq := emitChan(k, codec)

		seq := []T{gen(1), gen(2), gen(3)}
		for _, obj := range seq {
			snd <- obj
		}

		it.Then(t).Should(
			it.Equiv(<-dlq, seq[0]),
			it.Fail(func() error { return <-err }).Contain("lost"),

			it.Equiv(<-dlq, seq[1]),
			it.Fail(func() error { return <-err }).Contain("lost"),

			it.Equiv(<-dlq, seq[2]),
			it.Fail(func() error { return <-err }).Contain("lost"),
		)

		k.Close()
	})
}

//------------------------------------------------------------------------------

type devnil[T any] struct{}

func (e devnil[T]) Enq(ctx context.Context, bag swarm.Bag) error {
	return fmt.Errorf("lost")
}

func (e devnil[T]) Close() error { return nil }

func (e devnil[T]) Category() string { return "test" }

func (e devnil[T]) Encode(x T) (swarm.Bag, error) {
	return swarm.Bag{}, fmt.Errorf("invalid")
}
