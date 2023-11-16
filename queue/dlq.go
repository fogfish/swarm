//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package queue

import (
	"log/slog"

	"github.com/fogfish/swarm/internal/pipe"
)

// Consumes dead letter messages
//
// queue.LogDeadLetters(queue.Enqueue(...))
func LogDeadLetters[T any](out chan<- T, err <-chan T) chan<- T {
	pipe.ForEach[T](err, func(t T) {
		slog.Error("Fail to emit", "msg", t)
	})

	return out
}
