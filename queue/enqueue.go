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
	"log/slog"
	"reflect"
	"strings"

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

	cat := categoryOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	sock := q.Enqueue(cat, &ch)

	pipe.ForEach(ch.Msg, func(object T) {
		ch.Busy.Lock()
		defer ch.Busy.Unlock()

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

		slog.Debug("Enqueued message", "kind", "typed", "category", bag.Category, "object", object)
	})

	slog.Debug("Created enqueue channels: out, err", "kind", "typed", "category", cat)

	return ch.Msg, ch.Err
}

// normalized type name
func categoryOf[T any]() string {
	typ := reflect.TypeOf(new(T)).Elem()
	cat := typ.Name()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().Name()
	}

	seq := strings.Split(strings.Trim(cat, "]"), "[")
	tkn := make([]string, len(seq))
	for i, s := range seq {
		r := strings.Split(s, ".")
		tkn[i] = r[len(r)-1]
	}

	return strings.Join(tkn, "[") + strings.Repeat("]", len(tkn)-1)
}

func Must(broker swarm.Broker, err error) swarm.Broker {
	if err != nil {
		panic(err)
	}

	return broker
}
