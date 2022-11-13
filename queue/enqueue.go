//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package queue

import (
	"encoding/json"
	"reflect"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Enqueue creates pair of channels to send messages and dead-letter queue
*/
func Enqueue[T any](q swarm.Broker, category ...string) (chan<- T, <-chan T) {
	// TODO: discard dlq for At Most Once
	//       make it nil

	conf := q.Config()
	ch := swarm.NewMsgEnqCh[T](conf.EnqueueCapacity)

	cat := typeOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	sock := q.Enqueue(cat, ch)

	pipe.ForEach(ch.Msg, func(object T) {
		msg, err := json.Marshal(object)
		if err != nil {
			ch.Err <- object
			if conf.StdErr != nil {
				conf.StdErr <- err
			}
			return
		}

		bag := swarm.Bag{Category: cat, Object: msg}
		err = conf.Backoff.Retry(func() error { return sock.Enq(bag) })
		if err != nil {
			ch.Err <- object
			if conf.StdErr != nil {
				conf.StdErr <- err
			}
		}
	})

	return ch.Msg, ch.Err
}

func typeOf[T any]() string {
	//
	// TODO: fix
	//   Action[*swarm.User] if container type is used
	//

	typ := reflect.TypeOf(*new(T))
	cat := typ.Name()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().Name()
	}

	return cat
}

//
func Must(broker swarm.Broker, err error) swarm.Broker {
	if err != nil {
		panic(err)
	}

	return broker
}
