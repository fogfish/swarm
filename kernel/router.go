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

	"github.com/fogfish/golem/optics"
	"github.com/fogfish/swarm"
)

// Router is typed pair of message channel and codec
type msgRouter[T any] struct {
	ch    chan swarm.Msg[T]
	codec Decoder[T]
}

func newMsgRouter[T any](
	ch chan swarm.Msg[T],
	codec Decoder[T],
) msgRouter[T] {
	return msgRouter[T]{
		ch:    ch,
		codec: codec,
	}
}

func (a msgRouter[T]) Route(ctx context.Context, bag swarm.Bag) error {
	obj, err := a.codec.Decode(bag)
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

// Router is typed pair of message channel and codec
type evtRouter[E swarm.Event[M, T], M, T any] struct {
	ch    chan E
	codec Decoder[E]
	shape optics.Lens3[E, swarm.Digest, error, any]
}

func newEvtRouter[E swarm.Event[M, T], M, T any](
	ch chan E,
	codec Decoder[E],
) evtRouter[E, M, T] {
	return evtRouter[E, M, T]{
		ch:    ch,
		codec: codec,
		shape: optics.ForShape3[E, swarm.Digest, error, any](),
	}
}

func (a evtRouter[E, M, T]) Route(ctx context.Context, bag swarm.Bag) error {
	evt, err := a.codec.Decode(bag)
	if err != nil {
		slog.Debug("rouetr failed to decode event",
			slog.Any("cat", bag.Category),
			slog.Any("bag", bag),
			slog.Any("err", err),
		)
		return swarm.ErrDecoder.With(err)
	}

	a.shape.Put(&evt, bag.Digest, bag.Error, bag.IOContext)

	select {
	case <-ctx.Done():
		return swarm.ErrRouting.With(nil, bag.Category)
	case a.ch <- evt:
		return nil
	}
}
