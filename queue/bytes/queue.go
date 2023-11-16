//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package bytes

import (
	"log/slog"

	"github.com/fogfish/swarm"
)

type Queue interface {
	Enqueue([]byte) error
}

type queue struct {
	cat  string
	conf swarm.Config
	sock swarm.Enqueue
}

func (q queue) Sync()  {}
func (q queue) Close() {}

func (q queue) Enqueue(object []byte) error {
	bag := swarm.Bag{Category: q.cat, Object: object}
	err := q.conf.Backoff.Retry(func() error { return q.sock.Enq(bag) })
	if err != nil {
		return err
	}

	return nil
}

func New(q swarm.Broker, category string) Queue {
	queue := &queue{cat: category, conf: q.Config()}
	queue.sock = q.Enqueue(category, queue)

	slog.Debug("Created sync emitter", "kind", "bytes", "category", category)

	return queue
}
