//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package bytes

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// Dequeue bytes
func Dequeue(q swarm.Broker, cat string) (<-chan swarm.Msg[[]byte], chan<- swarm.Msg[[]byte]) {
	codec := swarm.NewCodecByte()

	return kernel.Dequeue(q.(*kernel.Kernel), cat, codec)
}
