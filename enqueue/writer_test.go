//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package enqueue_test

import (
	"context"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/enqueue"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

func TestNewTypes(t *testing.T) {
	mock := mockEmitter(10)
	k := kernel.NewEnqueuer(mock, swarm.Config{})

	q := enqueue.NewTyped[User](k)
	q.Enq(context.Background(), User{ID: "id", Text: "user"})

	it.Then(t).Should(
		it.Json(mock.val).Equiv(`{"id":"id","text":"user"}`),
	)

	k.Close()
}

func TestNewEvent(t *testing.T) {
	mock := mockEmitter(10)
	k := kernel.NewEnqueuer(mock, swarm.Config{})

	q := enqueue.NewEvent[swarm.Meta, User](k)
	q.Enq(context.Background(),
		swarm.Event[swarm.Meta, User]{
			Meta: &swarm.Meta{},
			Data: &User{ID: "id", Text: "user"},
		},
	)

	it.Then(t).Should(
		it.Json(mock.val).Equiv(`
			{
				"meta": {"type": "User", "id": "_", "created": "_"},
				"data": {"id":"id","text":"user"}
			}
		`),
	)

	k.Close()
}

func TestNewBytes(t *testing.T) {
	mock := mockEmitter(10)
	k := kernel.NewEnqueuer(mock, swarm.Config{})

	q := enqueue.NewBytes(k, encoding.NewCodecByte("User"))
	q.Enq(context.Background(), []byte(`{"id":"id","text":"user"}`))

	it.Then(t).Should(
		it.Json(mock.val).Equiv(`{"id":"id","text":"user"}`),
	)

	k.Close()
}
