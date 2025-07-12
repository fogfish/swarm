//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package listen

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

// Creates pair of channels to receive and acknowledge messages of type T
func Typed[T any](q *kernel.ListenerCore, codec ...kernel.Decoder[T]) (rcv <-chan swarm.Msg[T], ack chan<- swarm.Msg[T]) {
	var c kernel.Decoder[T]
	if len(codec) == 0 {
		c = encoding.ForTyped[T]()
	} else {
		c = codec[0]
	}

	return kernel.RecvChan(q, c.Category(), c)
}

// Creates pair of channels to receive and acknowledge events of type T
func Event[E swarm.Event[M, T], M, T any](q *kernel.ListenerCore, codec ...kernel.Decoder[swarm.Event[M, T]]) (<-chan swarm.Msg[swarm.Event[M, T]], chan<- swarm.Msg[swarm.Event[M, T]]) {
	var c kernel.Decoder[swarm.Event[M, T]]
	if len(codec) == 0 {
		c = encoding.ForEvent[E, M, T](q.Config.Realm, q.Config.Agent)
	} else {
		c = codec[0]
	}

	return kernel.RecvChan(q, c.Category(), c)
}

// Create pair of channels to receive and acknowledge pure binary
func Bytes(q *kernel.ListenerCore, codec kernel.Decoder[[]byte]) (<-chan swarm.Msg[[]byte], chan<- swarm.Msg[[]byte]) {
	return kernel.RecvChan(q, codec.Category(), codec)
}
