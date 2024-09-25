//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package queue

import (
	"context"
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// Enqueue message to the broker synchronously, blocking the routine until
// message is accepted by the broker.
type Enqueuer[T any] interface {
	Put(T) error
	Enq(string, T) error
}

type queue[T any] struct {
	cat   string
	codec kernel.Codec[T]
	emit  kernel.Emitter
}

func (q queue[T]) Put(object T) error { return q.Enq(q.cat, object) }

func (q queue[T]) Enq(cat string, object T) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	ctx := swarm.NewContext(context.Background(), cat, "")
	bag := swarm.Bag{Ctx: ctx, Object: msg}

	err = q.emit.Enq(bag)
	if err != nil {
		return err
	}

	slog.Debug("Enqueued message", "category", bag.Ctx.Category, "object", object)
	return nil
}

func New[T any](q swarm.Broker, category ...string) Enqueuer[T] {
	k := q.(*kernel.Kernel)

	cat := TypeOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	codec := swarm.NewCodecJson[T]()

	queue := &queue[T]{cat: cat, codec: codec, emit: k.Emitter}

	slog.Debug("Created sync emitter", "kind", "typed", "category", cat)

	return queue
}
