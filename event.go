//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"github.com/fogfish/curie"
	"github.com/fogfish/golem/pure"
)

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
	// It is automatically defined by the library upon the transmission
	ID string `json:"@id,omitempty"`

	//
	// Canonical IRI that defines a type of action.
	// It is automatically defined by the library upon the transmission
	Type curie.IRI `json:"@type,omitempty"`

	//
	// Direct performer of the event, a software service that emits action to the stream.
	Agent curie.IRI `json:"agent,omitempty"`

	//
	// Indirect participants, a user who initiated an event.
	Participant curie.IRI `json:"participant,omitempty"`

	//
	// ISO8601 timestamps when action has been created
	// It is automatically defined by the library upon the transmission
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
