//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package emit

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

// Creates pair of channels to emit messages of type T
func Typed[T any](q *kernel.EmitterIO, codec ...kernel.Encoder[T]) (snd chan<- T, dlq <-chan T) {
	var c kernel.Encoder[T]
	if len(codec) == 0 {
		c = encoding.ForTyped[T]()
	} else {
		c = codec[0]
	}

	return kernel.EmitChan(q, c)
}

// Creates pair of channels to emit events of type T
func Event[E swarm.Event[M, T], M, T any](q *kernel.EmitterIO, codec ...kernel.Encoder[swarm.Event[M, T]]) (snd chan<- swarm.Event[M, T], dlq <-chan swarm.Event[M, T]) {
	if len(codec) > 0 {
		return kernel.EmitEvent(q, codec[0])
	}

	c := encoding.ForEvent[E](q.Config.Realm, q.Config.Agent)
	return kernel.EmitEvent(q, c)
}

// Create pair of channels to emit pure binaries
func Bytes(q *kernel.EmitterIO, codec kernel.Encoder[[]byte]) (snd chan<- []byte, dlq <-chan []byte) {
	return kernel.EmitChan(q, codec)
}
