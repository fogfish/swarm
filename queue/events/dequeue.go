//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events

import (
	"encoding/json"
	"log/slog"

	"github.com/fogfish/golem/optics"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*
Dequeue ...
*/
func Dequeue[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) (<-chan *E, chan<- *E) {
	conf := q.Config()
	ch := swarm.NewEvtDeqCh[T, E](conf.DequeueCapacity)

	catE := categoryOf[E]()
	if len(category) > 0 {
		catE = category[0]
	}

	shape := optics.ForShape2[E, string, error]("Digest", "Err")

	sock := q.Dequeue(catE, &ch)

	pipe.ForEach(ch.Ack, func(object *E) {
		digest, fail := shape.Get(object)

		err := conf.Backoff.Retry(func() error {
			return sock.Ack(swarm.Bag{
				Category: catE,
				Digest:   digest,
				Err:      fail,
			})
		})
		if err != nil && conf.StdErr != nil {
			conf.StdErr <- err
		}
	})

	pipe.Emit(ch.Msg, q.Config().PollFrequency, func() (*E, error) {
		var bag swarm.Bag
		err := conf.Backoff.Retry(func() (err error) {
			bag, err = sock.Deq(catE)
			return
		})
		if err != nil {
			if conf.StdErr != nil {
				conf.StdErr <- err
			}
			return nil, err
		}

		evt := new(E)

		if bag.Event != nil {
			evt = bag.Event.(*E)
		}

		if bag.Object != nil {
			if err := json.Unmarshal(bag.Object, evt); err != nil {
				if conf.StdErr != nil {
					conf.StdErr <- err
				}
				return nil, err
			}
		}

		shape.Put(evt, bag.Digest, nil)

		return evt, nil
	})

	slog.Debug("Created dequeue channels: rcv, ack", "kind", "event", "category", catE)

	return ch.Msg, ch.Ack
}
