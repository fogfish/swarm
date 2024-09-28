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
func Type[T any](q *kernel.Dequeuer, category ...string) (rcv <-chan swarm.Msg[T], ack chan<- swarm.Msg[T]) {
	return kernel.Dequeue(q,
		swarm.TypeOf[T](category...),
		encoding.NewCodecJson[T](),
	)
}

// Creates pair of channels to receive and acknowledge events of type T
func Event[T any, E swarm.EventKind[T]](q *kernel.Dequeuer, category ...string) (<-chan swarm.Msg[*E], chan<- swarm.Msg[*E]) {
	cat := swarm.TypeOf[E](category...)

	return kernel.Dequeue(q, cat,
		encoding.NewCodecEvent[T, E](q.Config.Source, cat),
	)
}

// Create pair of channels to receive and acknowledge pure binary
func Bytes(q *kernel.Dequeuer, cat string) (<-chan swarm.Msg[[]byte], chan<- swarm.Msg[[]byte]) {
	if q.Config.Codec != nil {
		return kernel.Dequeue(q, cat, q.Config.Codec)
	}

	return kernel.Dequeue(q, cat, encoding.NewCodecByte())
}
