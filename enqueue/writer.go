//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package enqueue

import (
	"context"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

// Synchronous emitter of typed messages to the broker.
// It blocks the routine until the message is accepted by the broker.
type EmitterTyped[T any] struct {
	cat    string
	codec  kernel.Encoder[T]
	kernel *kernel.Enqueuer
}

// Creates synchronous typed emitter
func NewTyped[T any](q *kernel.Enqueuer, category ...string) *EmitterTyped[T] {
	return &EmitterTyped[T]{
		cat:    swarm.TypeOf[T](category...),
		codec:  encoding.NewCodecJson[T](),
		kernel: q,
	}
}

// Synchronously enqueue message to broker.
// It guarantees message to be send after return
func (q *EmitterTyped[T]) Enq(ctx context.Context, object T, cat ...string) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	category := q.cat
	if len(cat) > 0 {
		category = cat[0]
	}

	bag := swarm.Bag{
		Category: category,
		Object:   msg,
	}

	err = q.kernel.Emitter.Enq(ctx, bag)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

// Synchronous emitter of events to the broker.
// It blocks the routine until the event is accepted by the broker.
type EmitterEvent[M, T any] struct {
	cat    string
	codec  kernel.Encoder[swarm.Event[M, T]]
	kernel *kernel.Enqueuer
}

// Creates synchronous event emitter
func NewEvent[M, T any](q *kernel.Enqueuer, category ...string) *EmitterEvent[M, T] {
	cat := swarm.TypeOf[T](category...)

	return &EmitterEvent[M, T]{
		cat:    cat,
		codec:  encoding.NewCodecEvent[M, T](q.Config.Source, cat),
		kernel: q,
	}
}

// Synchronously enqueue event to broker.
// It guarantees event to be send after return.
func (q *EmitterEvent[M, T]) Enq(ctx context.Context, object swarm.Event[M, T], cat ...string) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	category := q.cat
	if len(cat) > 0 {
		category = cat[0]
	}

	bag := swarm.Bag{
		Category: category,
		Object:   msg,
	}

	err = q.kernel.Emitter.Enq(ctx, bag)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

// Synchronous emitter of raw packets (bytes) to the broker.
// It blocks the routine until the message is accepted by the broker.
type EmitterBytes struct {
	cat    string
	codec  kernel.Encoder[[]byte]
	kernel *kernel.Enqueuer
}

// Creates synchronous emitter
func NewBytes(q *kernel.Enqueuer, category string) *EmitterBytes {
	var codec swarm.Codec
	if q.Config.PacketCodec != nil {
		codec = q.Config.PacketCodec
	} else {
		codec = encoding.NewCodecByte()
	}

	return &EmitterBytes{
		cat:    category,
		codec:  codec,
		kernel: q,
	}
}

// Synchronously enqueue bytes to broker.
// It guarantees message to be send after return
func (q *EmitterBytes) Enq(ctx context.Context, object []byte, cat ...string) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	category := q.cat
	if len(cat) > 0 {
		category = cat[0]
	}

	bag := swarm.Bag{
		Category: category,
		Object:   msg,
	}

	err = q.kernel.Emitter.Enq(ctx, bag)
	if err != nil {
		return err
	}

	return nil
}
