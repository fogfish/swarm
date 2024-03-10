//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events

import (
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

type Queue[T any, E swarm.EventKind[T]] interface {
	Enqueue(*E) error
}

type queue[T any, E swarm.EventKind[T]] struct {
	cat   string
	codec kernel.Codec[*E]
	emit  kernel.Emitter
}

func (q queue[T, E]) Sync()  {}
func (q queue[T, E]) Close() {}

func (q queue[T, E]) Enqueue(object *E) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	bag := swarm.Bag{Category: q.cat, Object: msg}
	err = q.emit.Enq(bag)
	if err != nil {
		return err
	}

	slog.Debug("Enqueued event", "category", bag.Category, "object", object)
	return nil
}

func New[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) Queue[T, E] {
	k := q.(*kernel.Kernel)

	catE := categoryOf[E]()
	if len(category) > 0 {
		catE = category[0]
	}

	codec := swarm.NewCodecEvent[T, E](k.Config.Source, catE)

	queue := &queue[T, E]{
		cat:   catE,
		codec: codec,
		emit:  k.Emitter,
	}

	slog.Debug("Created sync emitter", "kind", "event", "category", catE)

	return queue
}
