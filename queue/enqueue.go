//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package queue

import (
	"reflect"
	"strings"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// Create egress and dead-letter queue channels for the category
func Enqueue[T any](q swarm.Broker, category ...string) (chan<- T, <-chan T) {
	cat := TypeOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	codec := swarm.NewCodecJson[T]()

	return kernel.Enqueue(q.(*kernel.Kernel), cat, codec)
}

// normalized type name
func TypeOf[T any]() string {
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
