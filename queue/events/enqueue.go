//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events

import (
	"reflect"
	"strings"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// Enqueue creates pair of channels
// - to send messages
// - failed messages (dead-letter queue)
func Enqueue[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) (chan<- *E, <-chan *E) {
	catE := categoryOf[E]()
	if len(category) > 0 {
		catE = category[0]
	}

	codec := swarm.NewCodecEvent[T, E]("TODO", catE)

	return kernel.Enqueue(q.(*kernel.Kernel), catE, codec)

	// panic("Not implemented")
	// conf := q.Config()
	// ch := swarm.NewEvtEnqCh[T, E](conf.EnqueueCapacity)

	// sock := q.Enqueue(catE, &ch)

	// pipe.ForEach(ch.Msg, func(object *E) {
	// 	ch.Busy.Lock()
	// 	defer ch.Busy.Unlock()

	// 	_, knd, src, _ := shape.Get(object)
	// 	if knd == "" {
	// 		knd = curie.IRI(catE)
	// 	}

	// 	if src == "" {
	// 		src = curie.IRI(q.Config().Source)
	// 	}

	// 	shape.Put(object, guid.G(guid.Clock).String(), knd, src, time.Now())

	// 	msg, err := json.Marshal(object)
	// 	if err != nil {
	// 		ch.Err <- object
	// 		if conf.StdErr != nil {
	// 			conf.StdErr <- err
	// 		}
	// 		return
	// 	}

	// 	bag := swarm.Bag{Category: catE, Object: msg}
	// 	err = conf.Backoff.Retry(func() error { return sock.Enq(bag) })
	// 	if err != nil {
	// 		ch.Err <- object
	// 		if conf.StdErr != nil {
	// 			conf.StdErr <- err
	// 		}
	// 	}

	// 	slog.Debug("Enqueued event", "kind", "event", "category", bag.Category, "object", object)
	// })

	// slog.Debug("Created enqueue channels: out, err", "kind", "event", "category", catE)

	// return ch.Msg, ch.Err
}

// normalized type name
func categoryOf[T any]() string {
	typ := reflect.TypeOf(new(T)).Elem()
	cat := typ.String()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().String()
	}

	seq := strings.Split(cat, "[")
	return seq[0]
}
