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
	"reflect"
	"strings"
	"time"

	"github.com/fogfish/curie"
	"github.com/fogfish/golem/optics"
	"github.com/fogfish/guid"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

// Enqueue creates pair of channels
// - to send messages
// - failed messages (dead-letter queue)
func Enqueue[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) (chan<- *E, <-chan *E) {
	conf := q.Config()
	ch := swarm.NewEvtEnqCh[T, E](conf.EnqueueCapacity)

	catT := strings.ToLower(categoryOf[T]())
	catE := categoryOf[E]()
	if len(category) > 0 {
		catE = category[0]
	}

	shape := optics.ForShape4[E, string, curie.IRI, curie.IRI, time.Time]("ID", "Type", "Agent", "Created")

	sock := q.Enqueue(catE, &ch)

	pipe.ForEach(ch.Msg, func(object *E) {
		ch.Busy.Lock()
		defer ch.Busy.Unlock()

		_, knd, src, _ := shape.Get(object)
		if knd == "" {
			knd = curie.IRI(catT + ":" + catE)
		}

		if src == "" {
			src = curie.IRI(q.Config().Source)
		}

		// TODO: migrate to v2
		shape.Put(object, guid.G.K(guid.Clock).String(), knd, src, time.Now())

		msg, err := json.Marshal(object)
		if err != nil {
			ch.Err <- object
			if conf.StdErr != nil {
				conf.StdErr <- err
			}
			return
		}

		bag := swarm.Bag{Category: catE, Object: msg}
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
