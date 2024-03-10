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
	catE := categoryOf[E]()
	if len(category) > 0 {
		catE = category[0]
	}

	codec := swarm.NewCodecEvent[T, E]("TODO", catE)

	return kernel.Dequeue(q.(*kernel.Kernel), catE, codec)

	// panic("Not implemented")
	// conf := q.Config()
	// ch := swarm.NewEvtDeqCh[T, E](conf.DequeueCapacity)

	// lens := optics.ForProduct1[E, swarm.Digest]("Digest")

	// sock := q.Dequeue(catE, &ch)

	// pipe.ForEach(ch.Ack, func(object *E) {
	// 	digest := lens.Get(object)

	// 	err := conf.Backoff.Retry(func() error {
	// 		return sock.Ack(swarm.Bag{
	// 			Category: catE,
	// 			Digest:   digest,
	// 		})
	// 	})
	// 	if err != nil && conf.StdErr != nil {
	// 		conf.StdErr <- err
	// 		return
	// 	}

	// 	slog.Debug("Broker ack'ed object", "kind", "event", "category", catE, "object", object, "error", digest.Error)
	// })

	// pipe.Emit(ch.Msg, q.Config().PollFrequency, func() (*E, error) {
	// 	var bag swarm.Bag
	// 	err := conf.Backoff.Retry(func() (err error) {
	// 		bag, err = sock.Deq(catE)
	// 		return
	// 	})
	// 	if err != nil {
	// 		if conf.StdErr != nil {
	// 			conf.StdErr <- err
	// 		}
	// 		return nil, err
	// 	}

	// 	evt := new(E)

	// 	if bag.Event != nil {
	// 		evt = bag.Event.(*E)
	// 	}

	// 	if bag.Object != nil {
	// 		if err := json.Unmarshal(bag.Object, evt); err != nil {
	// 			if conf.StdErr != nil {
	// 				conf.StdErr <- err
	// 			}
	// 			return nil, err
	// 		}
	// 	}

	// 	lens.Put(evt, bag.Digest)

	// 	slog.Debug("Broker received object", "kind", "event", "category", catE, "object", evt)

	// 	return evt, nil
	// })

	// slog.Debug("Created dequeue channels: rcv, ack", "kind", "event", "category", catE)

	// return ch.Msg, ch.Ack
}
