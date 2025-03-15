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
	"log/slog"

	"github.com/fogfish/swarm"
)

// Router is typed pair of message channel and codec
type router[T any] struct {
	ch    chan swarm.Msg[T]
	codec Decoder[T]
}

func (a router[T]) Route(ctx context.Context, bag swarm.Bag) error {
	obj, err := a.codec.Decode(bag.Object)
	if err != nil {
		slog.Debug("rouetr failed to decode message",
			slog.Any("cat", bag.Category),
			slog.Any("bag", bag),
			slog.Any("err", err),
		)
		return swarm.ErrDecoder.With(err)
	}

	msg := swarm.ToMsg(bag, obj)

	select {
	case <-ctx.Done():
		return swarm.ErrRouting.With(nil, bag.Category)
	case a.ch <- msg:
		return nil
	}
}
