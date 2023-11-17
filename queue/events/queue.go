//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/fogfish/curie"
	"github.com/fogfish/golem/optics"
	"github.com/fogfish/guid/v2"
	"github.com/fogfish/swarm"
)

type Queue[T any, E swarm.EventKind[T]] interface {
	Enqueue(*E) error
}

type queue[T any, E swarm.EventKind[T]] struct {
	cat   string
	conf  swarm.Config
	sock  swarm.Enqueue
	shape optics.Lens4[E, string, curie.IRI, curie.IRI, time.Time]
}

func (q queue[T, E]) Sync()  {}
func (q queue[T, E]) Close() {}

func (q queue[T, E]) Enqueue(object *E) error {
	_, knd, src, _ := q.shape.Get(object)
	if knd == "" {
		knd = curie.IRI(q.cat)
	}

	if src == "" {
		src = curie.IRI(q.conf.Source)
	}

	q.shape.Put(object, guid.G(guid.Clock).String(), knd, src, time.Now())

	msg, err := json.Marshal(object)
	if err != nil {
		return err
	}

	bag := swarm.Bag{Category: q.cat, Object: msg}
	err = q.conf.Backoff.Retry(func() error { return q.sock.Enq(bag) })
	if err != nil {
		return err
	}

	slog.Debug("Enqueued event", "category", bag.Category, "object", object)
	return nil
}

func New[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) Queue[T, E] {
	catE := categoryOf[E]()
	if len(category) > 0 {
		catE = category[0]
	}

	shape := optics.ForShape4[E, string, curie.IRI, curie.IRI, time.Time]("ID", "Type", "Agent", "Created")

	queue := &queue[T, E]{
		cat:   catE,
		conf:  q.Config(),
		shape: shape,
	}
	queue.sock = q.Enqueue(catE, queue)

	slog.Debug("Created sync emitter", "kind", "event", "category", catE)

	return queue
}
