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
	"github.com/fogfish/swarm/internal/kernel"
)

type Queue interface {
	Enqueue([]byte) error
}

type queue struct {
	cat   string
	codec kernel.Codec[[]byte]
	emit  kernel.Emitter
}

func (q queue) Sync()  {}
func (q queue) Close() {}

func (q queue) Enqueue(object []byte) error {
	bag := swarm.Bag{Category: q.cat, Object: object}
	err := q.emit.Enq(bag)
	if err != nil {
		return err
	}

	slog.Debug("Enqueued bytes", "category", bag.Category, "object", object)
	return nil
}

func New(q swarm.Broker, category string) Queue {
	k := q.(*kernel.Kernel)

	codec := swarm.NewCodecByte()

	queue := &queue{cat: category, codec: codec, emit: k.Emitter}

	slog.Debug("Created sync emitter", "kind", "bytes", "category", category)

	return queue
}
