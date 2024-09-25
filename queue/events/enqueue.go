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
	"github.com/fogfish/swarm/kernel"
)

// Enqueue creates pair of channels
// - to send messages
// - failed messages (dead-letter queue)
func Enqueue[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) (chan<- *E, <-chan *E) {
	catE := TypeOf[E]()
	if len(category) > 0 {
		catE = category[0]
	}

	k := q.(*kernel.Kernel)
	codec := swarm.NewCodecEvent[T, E](k.Config.Source, catE)

	return kernel.Enqueue(q.(*kernel.Kernel), catE, codec)
}

// normalized type name
func TypeOf[T any]() string {
	typ := reflect.TypeOf(new(T)).Elem()
	cat := typ.String()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().String()
	}

	seq := strings.Split(cat, "[")
	return seq[0]
}
