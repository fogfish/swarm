//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package broadcast_test

import (
	"context"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm/kernel/broadcast"
)

func TestBroadcast(t *testing.T) {
	b := broadcast.New()

	worker := func() {
		ch := b.Register()

		go func() {
			defer b.Unregister(ch)
			(<-ch) <- struct{}{}
		}()
	}

	t.Run("Cast.1", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		worker()
		err := b.Cast(ctx)

		it.Then(t).Should(it.Nil(err))
	})

	t.Run("Cast.N", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		worker()
		worker()
		worker()
		err := b.Cast(ctx)

		it.Then(t).Should(it.Nil(err))
	})

	t.Run("Timeout.1", func(t *testing.T) {
		ch := b.Register()
		defer b.Unregister(ch)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		it.Then(t).Should(
			it.Fail(
				func() error {
					return b.Cast(ctx)
				},
			).Contain("context deadline exceeded"),
		)
	})

	t.Run("Timeout.N", func(t *testing.T) {
		worker()
		worker()
		worker()

		ch := b.Register()
		defer b.Unregister(ch)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		it.Then(t).Should(
			it.Fail(
				func() error {
					return b.Cast(ctx)
				},
			).Contain("context deadline exceeded"),
		)
	})

	t.Run("Cancel.1", func(t *testing.T) {
		ch := b.Register()
		defer b.Unregister(ch)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		it.Then(t).Should(
			it.Fail(
				func() error {
					return b.Cast(ctx)
				},
			).Contain("context canceled"),
		)
	})
}
