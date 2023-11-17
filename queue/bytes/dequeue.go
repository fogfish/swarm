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

// Dequeue bytes
func Dequeue(q swarm.Broker, cat string) (<-chan *swarm.Msg[[]byte], chan<- *swarm.Msg[[]byte]) {
	conf := q.Config()
	ch := swarm.NewMsgDeqCh[[]byte](conf.DequeueCapacity)

	sock := q.Dequeue(cat, &ch)

	pipe.ForEach(ch.Ack, func(object *swarm.Msg[[]byte]) {
		err := conf.Backoff.Retry(func() error {
			return sock.Ack(swarm.Bag{
				Category: cat,
				Digest:   object.Digest,
			})
		})
		if err != nil && conf.StdErr != nil {
			conf.StdErr <- err
			return
		}

		slog.Debug("Broker ack'ed object", "kind", "bytes", "category", cat, "object", object.Object, "error", object.Digest.Error)
	})

	pipe.Emit(ch.Msg, q.Config().PollFrequency, func() (*swarm.Msg[[]byte], error) {
		var bag swarm.Bag
		err := conf.Backoff.Retry(func() (err error) {
			bag, err = sock.Deq(cat)
			return
		})
		if err != nil {
			if conf.StdErr != nil {
				conf.StdErr <- err
			}
			return nil, err
		}

		msg := &swarm.Msg[[]byte]{
			Object: bag.Object,
			Digest: bag.Digest,
		}

		slog.Debug("Broker received object", "kind", "bytes", "category", cat, "object", bag.Object)

		return msg, nil
	})

	slog.Debug("Created dequeue channels: rcv, ack", "kind", "bytes", "category", cat)

	return ch.Msg, ch.Ack
}
