//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package bytes

import (
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

// Enqueue creates pair of channels to send messages and dead-letter queue
func Enqueue(q swarm.Broker, cat string) (chan<- []byte, <-chan []byte) {
	conf := q.Config()
	ch := swarm.NewMsgEnqCh[[]byte](conf.EnqueueCapacity)

	sock := q.Enqueue(cat, &ch)

	pipe.ForEach(ch.Msg, func(object []byte) {
		ch.Busy.Lock()
		defer ch.Busy.Unlock()

		bag := swarm.Bag{Category: cat, Object: object}
		err := conf.Backoff.Retry(func() error { return sock.Enq(bag) })
		if err != nil {
			ch.Err <- object
			if conf.StdErr != nil {
				conf.StdErr <- err
			}
		}

		slog.Debug("Enqueued", "kind", "bytes", "category", bag.Category, "object", object)
	})

	slog.Debug("Created enqueue channels: out, err", "kind", "bytes", "category", cat)

	return ch.Msg, ch.Err
}
