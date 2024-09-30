//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package dequeue

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

// Creates pair of channels to receive and acknowledge messages of type T
func Typed[T any](q *kernel.Dequeuer, category ...string) (rcv <-chan swarm.Msg[T], ack chan<- swarm.Msg[T]) {
	return kernel.Dequeue(q,
		swarm.TypeOf[T](category...),
		encoding.NewCodecJson[T](),
	)
}

// Creates pair of channels to receive and acknowledge events of type T
func Event[M, T any](q *kernel.Dequeuer, category ...string) (<-chan swarm.Msg[swarm.Event[M, T]], chan<- swarm.Msg[swarm.Event[M, T]]) {
	cat := swarm.TypeOf[T](category...)

	return kernel.Dequeue(q, cat,
		encoding.NewCodecEvent[M, T](q.Config.Source, cat),
	)
}

// Create pair of channels to receive and acknowledge pure binary
func Bytes(q *kernel.Dequeuer, cat string) (<-chan swarm.Msg[[]byte], chan<- swarm.Msg[[]byte]) {
	if q.Config.PacketCodec != nil {
		return kernel.Dequeue(q, cat, q.Config.PacketCodec)
	}

	return kernel.Dequeue(q, cat, encoding.NewCodecByte())
}
