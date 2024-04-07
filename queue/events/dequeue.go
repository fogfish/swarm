//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// Dequeue event
func Dequeue[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) (<-chan swarm.Msg[*E], chan<- swarm.Msg[*E]) {
	catE := TypeOf[E]()
	if len(category) > 0 {
		catE = category[0]
	}

	codec := swarm.NewCodecEvent[T, E]("TODO", catE)

	return kernel.Dequeue(q.(*kernel.Kernel), catE, codec)
}
