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
	cat := categoryOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	codec := swarm.NewCodecJson[T]()

	return kernel.Dequeue(q.(*kernel.Kernel), cat, codec)

	// // TODO: automatically ack At Most Once, no ack channel
	// //       make it as /dev/null
	// conf := q.Config()
	// ch := swarm.NewMsgDeqCh[T](conf.DequeueCapacity)

	// sock := q.Dequeue(cat, &ch)

	// pipe.ForEach(ch.Ack, func(object *swarm.Msg[T]) {
	// 	err := conf.Backoff.Retry(func() error {
	// 		return sock.Ack(swarm.Bag{
	// 			Category: cat,
	// 			Digest:   object.Digest,
	// 		})
	// 	})
	// 	if err != nil && conf.StdErr != nil {
	// 		conf.StdErr <- err
	// 		return
	// 	}

	// 	slog.Debug("Broker ack'ed object", "kind", "typed", "category", cat, "object", object.Object, "error", object.Digest.Error)
	// })

	// pipe.Emit(ch.Msg, q.Config().PollFrequency, func() (*swarm.Msg[T], error) {
	// 	var bag swarm.Bag
	// 	err := conf.Backoff.Retry(func() (err error) {
	// 		bag, err = sock.Deq(cat)
	// 		return
	// 	})
	// 	if err != nil {
	// 		if conf.StdErr != nil {
	// 			conf.StdErr <- err
	// 		}
	// 		return nil, err
	// 	}

	// 	msg := &swarm.Msg[T]{Digest: bag.Digest}
	// 	if err := json.Unmarshal(bag.Object, &msg.Object); err != nil {
	// 		if conf.StdErr != nil {
	// 			conf.StdErr <- err
	// 		}
	// 		return nil, err
	// 	}

	// 	slog.Debug("Broker received object", "kind", "typed", "category", cat, "object", msg.Object)

	// 	return msg, nil
	// })

	// slog.Debug("Created dequeue channels: rcv, ack", "kind", "typed", "category", cat)

	// return ch.Msg, ch.Ack
}
