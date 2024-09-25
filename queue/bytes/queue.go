//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package bytes

import (
	"context"
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

type Enqueuer interface {
	Put([]byte) error
	Enq(string, []byte) error
}

type queue struct {
	cat   string
	codec kernel.Codec[[]byte]
	emit  kernel.Emitter
}

func (q queue) Put(object []byte) error { return q.Enq(q.cat, object) }

func (q queue) Enq(cat string, object []byte) error {
	obj, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	ctx := swarm.NewContext(context.Background(), cat, "")
	bag := swarm.Bag{Ctx: ctx, Object: obj}

	err = q.emit.Enq(bag)
	if err != nil {
		return err
	}

	slog.Debug("Enqueued bytes", "category", bag.Ctx.Category, "object", object)
	return nil
}

func New(q swarm.Broker, category string) Enqueuer {
	k := q.(*kernel.Kernel)

	codec := swarm.NewCodecByte()

	queue := &queue{cat: category, codec: codec, emit: k.Emitter}

	slog.Debug("Created sync emitter", "kind", "bytes", "category", category)

	return queue
}
