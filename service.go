//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"log/slog"
)

// Message broker
type Broker interface {
	Close()
	Await()
}

// Consumes dead letter messages
//
// swarm.LogDeadLetters(queue.Enqueue(...))
func LogDeadLetters[T any](out chan<- T, err <-chan T) chan<- T {
	go func() {
		var x T

		for x = range err {
			slog.Error("Fail to emit", "object", x)
		}
	}()

	return out
}
