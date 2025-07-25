//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"reflect"
	"strings"
)

// Unique brief summary of the message, specific to the broker
type Digest string

// Msg is a generic envelop type for incoming messages.
// It contains both decoded object and its digest used to acknowledge message.
type Msg[T any] struct {
	// Message category ~ topic
	Category string

	// Unique brief summary of the message
	Digest Digest

	// Error on the message processing
	Error error

	// I/O Context of the message, as obtained from broker
	IOContext any

	// Message decoded content
	Object T
}

// Fail message with error
func (msg Msg[T]) Fail(err error) Msg[T] {
	msg.Error = err
	return msg
}

// Bag is an abstract container for octet stream.
// Bag is used by the transport to abstract message on the wire.
type Bag struct {
	// Message category ~ topic
	Category string

	// Unique brief summary of the message
	Digest Digest

	// Error on the message processing
	Error error

	// I/O Context of the message, as obtained from broker
	IOContext any

	// Message raw content
	Object []byte
}

func ToMsg[T any](bag Bag, object T) Msg[T] {
	return Msg[T]{
		Category:  bag.Category,
		Digest:    bag.Digest,
		Error:     bag.Error,
		IOContext: bag.IOContext,
		Object:    object,
	}
}

// func ToBag[T any](msg Msg[T], object []byte) Bag {
// 	return Bag{
// 		Category: msg.Category,
// 		Digest:   msg.Digest,
// 		Error:    msg.Error,
// 		Object:   object,
// 	}
// }

// TypeOf returns normalized name of the type T.
func TypeOf[T any](category ...string) string {
	if len(category) > 0 {
		return category[0]
	}

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
