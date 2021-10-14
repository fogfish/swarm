//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"context"
	"sync"
	"time"
)

/*

Category of message (aka subject or topic), each message has unique type
*/
type Category string

/*

Msg type is an abstract container for octet stream, exchanged via channels

  ch <- swarm.Bytes("some message")
  ...
	for msg := range ch {
		msg.Bytes()
	}
*/
type Msg interface {
	Bytes() []byte
}

/*

Bytes (octet stream) is a built in type to represent sequence of bytes
as message.

  ch <- swarm.Bytes("some message")

*/
type Bytes []byte

/*

Bytes returns message payload (octet stream)
*/
func (b Bytes) Bytes() []byte { return b }

/*

Queue ...
*/
type Queue interface {
	/*
		Swarm System of Queue
	*/
	Sys() System

	/*
		Creates endpoints to receive messages and acknowledge its consumption.
	*/
	Recv(Category) (<-chan Msg, chan<- Msg)

	/*
		Creates endpoints to send messages and channel to consume errors.
	*/
	Send(Category) (chan<- Msg, <-chan Msg)

	/*
		Wait queue until all messages are flushed. Waiting is required in
		serverless environment just before lambda going to sleep.
	*/
	Wait()

	// TODO:
	// - Err (chan<- error) handle transport errors
	// - consider send failure as transport error (coupled design vs generic)
	// - consider ack as additional channel (?)
	// - consider Listen() <- chan error
}

/*

System ...
*/
type System interface {

	/*
		Unique System ID
	*/
	ID() string

	/*
		Spawn go routine in context of system
	*/
	Go(func(context.Context))

	/*
	 Stop system and all active go routines
	*/
	Stop()

	/*
	 Wait system to be stopped
	*/
	Wait()
}

type system struct {
	sync.Mutex

	id      string
	context context.Context
	cancel  context.CancelFunc
}

/*

New creates new queueing system
*/
func New(id string, opts ...Config) System {
	sys := &system{id: id, context: context.Background()}

	for _, opt := range opts {
		opt(sys)
	}

	sys.context, sys.cancel = context.WithCancel(sys.context)
	return sys
}

/*

Config of System Type
*/
type Config func(sys *system)

/*

WithContext config system with custom context
*/
func WithContext(ctx context.Context) Config {
	return func(sys *system) {
		sys.context = ctx
	}
}

var (
	_ System = (*system)(nil)
)

/*

ID return unique system ID
*/
func (sys *system) ID() string {
	return sys.id
}

/*

Spawn go routine in context of system
*/
func (sys *system) Go(f func(context.Context)) {
	go f(sys.context)
}

/*

Stop ...
*/
func (sys *system) Stop() {
	// TODO: use event based approach to control shutdown
	time.Sleep(5 * time.Second)
	sys.cancel()
	time.Sleep(5 * time.Second)
}

/*

Wait ...
*/
func (sys *system) Wait() {
	<-sys.context.Done()
}

/*

Must ensures successful creation of queue
*/
func Must(q Queue, err error) Queue {
	if err != nil {
		panic(err)
	}

	return q
}
