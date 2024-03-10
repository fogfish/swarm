//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

type Digest struct {
	// Unique brief summary of the message
	Brief string

	// Error on the message processing
	Error error
}

// Msg is a generic envelop type for incoming messages.
// It contains both decoded object and its digest used to acknowledge message.
type Msg[T any] struct {
	Object T
	Digest Digest
}

// Fail message with error
func (msg *Msg[T]) Fail(err error) *Msg[T] {
	msg.Digest.Error = err
	return msg
}

// Bag is an abstract container for octet stream.
// Bag is used by the transport to abstract message on the wire.
type Bag struct {
	Category string
	Object   []byte
	Digest   Digest
}
