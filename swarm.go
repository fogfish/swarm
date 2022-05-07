//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

/*

Msg type is an abstract container for octet stream, exchanged via channels

  ch <- swarm.Bytes("some message")
  ...
	for msg := range ch {
		msg.Bytes()
	}
*/
type Object interface {
	Bytes() []byte
}

/*

Msg type defines external ingress message.
It containers both payload and receipt to acknowledge
*/
type Msg struct {
	Payload []byte
	Receipt string
}

var (
	_ Object = (*Msg)(nil)
)

/*

Bytes returns message payload (octet stream)
*/
func (msg *Msg) Bytes() []byte {
	return msg.Payload
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

Bag is an internal message envelop containing message and routing attributes

TODO: Identity id
 - System
 - Queue
 - Category (type)
 - ==> Or system:queue/category

*/
type Bag struct {
	System   string
	Queue    string
	Category string

	// message payload
	Object Object

	//
	StdErr chan<- Object
}

type Sender interface {
	ID() string
	Start() error
	Close() error
	Send() chan *Bag
}

type Recver interface {
	ID() string
	Start() error
	Close() error
	Recv() chan *Bag
	Conf() chan *Bag
}

/*

EventBus is an an abstraction of transport protocol(s) to send & recv events
*/
type EventBus interface {
	ID() string

	// Send connects to queueing broker and returns channel to send messages
	Send() (chan *Bag, error)

	// Recv connects to queueing broker and returns channel to recv messages
	Recv() (chan *Bag, error)

	// Conf connects to queueing broker and returns channel to confirm processed messages
	Conf() (chan *Bag, error)
}

/*

Queue ...
*/
type Queue interface {
	// Creates endpoints to receive messages and acknowledge its consumption.
	Recv(string) (<-chan Object, chan<- Object)

	// Creates endpoints to send messages and channel to consume errors.
	Send(string) (chan<- Object, <-chan Object)

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
	// Queue creates new queuing endpoint
	Queue(Sender, Recver) Queue

	// Listen ...
	Listen() error

	/*
	 Stop system and all active go routines
	*/
	Stop()

	/*
	 Wait system to be stopped
	*/
	Wait()
}
