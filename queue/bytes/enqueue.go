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

// Enqueue creates pair of channels to send messages and dead-letter queue
func Enqueue(q swarm.Broker, cat string) (chan<- []byte, <-chan []byte) {
	codec := swarm.NewCodecByte()

	return kernel.Enqueue(q.(*kernel.Kernel), cat, codec)
}
