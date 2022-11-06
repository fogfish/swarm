//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"time"

	"github.com/fogfish/curie"
	"github.com/fogfish/golem/pure"
)

/*

Msg is a generic envelop type for incoming messages.
It contains both decoded object and its digest used to acknowledge message.
*/
type Msg[T any] struct {
	Object T
	Digest string
}

type EventType any

type EventKind[A any] pure.HKT[EventType, A]

/*

Event defines immutable fact(s) placed into the queueing system.
Event resembles the concept of Action as it is defined by schema.org.

  > An action performed by a direct agent and indirect participants upon a direct object.

This type supports development of event-driven solutions that treat data as
a collection of immutable facts, which are queried and processed in real-time.
These applications processes logical log of events, each event defines a change
to current state of the object, i.e. which attributes were inserted,
updated or deleted (a kind of diff). The event identifies the object that was
changed together with  using unique identifier.
*/
type Event[T any] struct {
	//
	// Unique identity for event
	ID string `json:"@id,omitempty"`

	//
	// Canonical IRI that defines a type of action.
	Type curie.IRI `json:"@type,omitempty"`

	//
	// Direct performer of the event, a software service that emits action to the stream.
	Agent curie.IRI `json:"agent,omitempty"`

	//
	// Indirect participants, a user who initiated an event.
	Participant curie.IRI `json:"participant,omitempty"`

	//
	// ISO8601 timestamps when action has been created
	Created string `json:"created,omitempty"`

	//
	// The digest of received event (used internally to ack processing)
	Digest string `json:"-"`

	//
	// The object upon which the event is carried out.
	Object T `json:"object,omitempty"`
}

func (Event[T]) HKT1(EventType) {}
func (Event[T]) HKT2(T)         {}

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

type Closer interface {
	Sync()
	Close()
}

/*

MsgSendCh is the pair of channel, exposed by the queue to clients to send messages
*/
type MsgSendCh[T any] struct {
	Msg chan T // channel to send message out
	Err chan T // channel to recv failed messages
}

func (ch MsgSendCh[T]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Err) == 0 {
			break
		}
	}
}

func (ch MsgSendCh[T]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Err)
}

/*

msgRecv is the pair of channel, exposed by the queue to clients to recv messages
*/
type MsgRecvCh[T any] struct {
	Msg chan *Msg[T] // channel to recv message
	Ack chan *Msg[T] // channel to send acknowledgement
}

func (ch MsgRecvCh[T]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Ack) == 0 {
			break
		}
	}
}

func (ch MsgRecvCh[T]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Ack)
}
