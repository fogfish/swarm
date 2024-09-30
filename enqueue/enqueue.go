//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package enqueue

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

// Creates pair of channels to emit messages of type T
func Typed[T any](q *kernel.Enqueuer, category ...string) (snd chan<- T, dlq <-chan T) {
	return kernel.Enqueue(q,
		swarm.TypeOf[T](category...),
		encoding.NewCodecJson[T](),
	)
}

// Creates pair of channels to emit events of type T
func Event[M, T any](q *kernel.Enqueuer, category ...string) (snd chan<- swarm.Event[M, T], dlq <-chan swarm.Event[M, T]) {
	cat := swarm.TypeOf[T](category...)

	return kernel.Enqueue(q, cat,
		encoding.NewCodecEvent[M, T](q.Config.Source, cat),
	)
}

// Create pair of channels to emit pure binaries
func Bytes(q *kernel.Enqueuer, cat string) (snd chan<- []byte, dlq <-chan []byte) {
	if q.Config.PacketCodec != nil {
		return kernel.Enqueue(q, cat, q.Config.PacketCodec)
	}

	return kernel.Enqueue(q, cat, encoding.NewCodecByte())
}
