//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

/*

Msg is a generic envelop type for incoming messages.
It contains both decoded object and its digest used to acknowledge message.
*/
type Msg[T any] struct {
	Object T
	Digest string
}

/*

Bag type is an abstract container for octet stream, exchanged via channels.
It is used by queuing transport to abstract message on the wire.
*/
type Bag struct {
	System   string
	Queue    string
	Category string

	Object []byte
	Digest string
}

/*

BagStdErr is an envelop used to enqueue messages to broker
*/
type BagStdErr struct {
	Bag
	StdErr func(error)
}

/*

Enqueue is an abstraction of transport protocols to send messages for
the processing to queueing system (broker). Enqueue is an abstraction of
transport client.
*/
type Enqueue interface {
	// Listent activates transport
	Listen() error

	// Close transport
	Close() error

	// Enq returns a channel to enqueue messages
	Enq() chan *BagStdErr

	// Enqueue synchronously message
	EnqSync(*Bag) error
}

/*

Dequeue is an abstraction of transport protocol to receive messages from
queueing system (broker). Dequeue is an abstraction of transport client.
*/
type Dequeue interface {
	// Listent activates transport
	Listen() error

	// Close transport
	Close() error

	// Deq returns a channel to dequeue messages
	Deq() chan *Bag

	// Ack returns a channel to acknowledge message processign
	Ack() chan *Bag
}

/*

Queue ...
*/
type Queue interface {
	// Sync queue buffers
	Sync()
}

/*

System of Queues controls group related queues.
*/
type System interface {
	ID() string

	// Create new queueing endpoint from transport
	Queue(string, Enqueue, Dequeue, *Policy) Queue

	// Listent activates all transport
	Listen() error

	// Close system and all queues
	Close()

	// Wait for system to be stopped
	Wait()
}
