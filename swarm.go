//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

/*

Messages
 * Client
   * send: T
   * recv: T + receipt

 * Transport
   * send: bytes, message attributes, recovery function
	 * recv: binary + receipt using codec function
	 * conf: receipt only
*/

/*

Msg type is an abstract container for octet stream, exchanged via channels

  ch <- swarm.Bytes("some message")
  ...
	for msg := range ch {
		msg.Bytes()
	}
*/
// type Object interface {
// 	Bytes() []byte
// }

/*

Msg type defines external ingress message.
It containers both payload and receipt to acknowledge
*/
// type Msg struct {
// 	Payload []byte
// 	Receipt string
// }

// var (
// 	_ Object = (*Msg)(nil)
// )

/*

Bytes returns message payload (octet stream)
*/
// func (msg *Msg) Bytes() []byte {
// 	return msg.Payload
// }

/*

Bytes (octet stream) is a built in type to represent sequence of bytes
as message.

  ch <- swarm.Bytes("some message")

*/
// type Bytes []byte

/*

Bytes returns message payload (octet stream)
*/
// func (b Bytes) Bytes() []byte { return b }

/*

Msg is a generic envelop type for incoming messages.
It contains both decoded object and its digest used to acknowledge message.
*/
type Msg[T any] struct {
	Object T
	Digest string
}

/*

Bytes is an abstract container for octet stream for incoming messages.
*/
type Bytes struct {
	Object []byte
	Digest string
}

//
// Enq
// Deq
//
// Enqueue
// Dequeue
//

// TODO:
//  [x] fix bag type
//  [x] fix recv dispatch
//  [x] finish generic Send (Enq) / Recv (Deq)
//  [x] fix MsgG to Msg[T]{Object, Digest}
//  - make bytes Send (Enq) / Recv (Deq)
//  [x] rename to enqueue

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
	// Creates endpoints to receive messages and acknowledge its consumption.
	// Recv(string) (<-chan Object, chan<- Object)

	// Creates endpoints to send messages and channel to consume errors.
	// Send(string) (chan<- Object, <-chan Object)

	// TODO:
	// - Err (chan<- error) handle transport errors
	// - consider send failure as transport error (coupled design vs generic)
	// - consider ack as additional channel (?)
	// - consider Listen() <- chan error
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
