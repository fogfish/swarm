//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package queue

import (
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

type Queue[T any] interface {
	Enqueue(T) error
}

type queue[T any] struct {
	cat   string
	codec kernel.Codec[T]
	emit  kernel.Emitter
}

func (q queue[T]) Sync()  {}
func (q queue[T]) Close() {}

func (q queue[T]) Enqueue(object T) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	bag := swarm.Bag{Category: q.cat, Object: msg}
	err = q.emit.Enq(bag)
	if err != nil {
		return err
	}

	slog.Debug("Enqueued message", "category", bag.Category, "object", object)
	return nil
}

func New[T any](q swarm.Broker, category ...string) Queue[T] {
	k := q.(*kernel.Kernel)

	cat := categoryOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	codec := swarm.NewCodecJson[T]()

	queue := &queue[T]{cat: cat, codec: codec, emit: k.Emitter}

	slog.Debug("Created sync emitter", "kind", "typed", "category", cat)

	return queue
}
