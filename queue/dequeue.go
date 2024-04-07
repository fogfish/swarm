//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package queue

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// Dequeue message
func Dequeue[T any](q swarm.Broker, category ...string) (<-chan swarm.Msg[T], chan<- swarm.Msg[T]) {
	cat := TypeOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	codec := swarm.NewCodecJson[T]()

	return kernel.Dequeue(q.(*kernel.Kernel), cat, codec)
}
