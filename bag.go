//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import "context"

// Global context for the message
type Context struct {
	context.Context

	// Message category ~ topic
	Category string

	// Unique brief summary of the message
	Digest string

	// Error on the message processing
	Error error
}

func NewContext(ctx context.Context, cat, digest string) *Context {
	return &Context{
		Context:  ctx,
		Category: cat,
		Digest:   digest,
	}
}

// type Digest struct {
// 	// Unique brief summary of the message
// 	Brief string

// 	// Error on the message processing
// 	Error error
// }

// Msg is a generic envelop type for incoming messages.
// It contains both decoded object and its digest used to acknowledge message.
type Msg[T any] struct {
	Ctx    *Context
	Object T
	// Digest Digest
}

// Fail message with error
func (msg Msg[T]) Fail(err error) Msg[T] {
	msg.Ctx.Error = err
	//msg.Digest.Error = err
	return msg
}

// Bag is an abstract container for octet stream.
// Bag is used by the transport to abstract message on the wire.
type Bag struct {
	Ctx    *Context
	Object []byte
	// Digest   Digest
	// Category string
}
