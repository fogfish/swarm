//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"time"

	"github.com/fogfish/curie/v2"
)

// Event defines immutable fact(s) placed into the queueing system.
// Event resembles the concept of Action as it is defined by schema.org.
//
//	> An action performed by a direct agent and indirect participants upon a direct object.
//
// This type supports development of event-driven solutions that treat data as
// a collection of immutable facts, which are queried and processed in real-time.
// These applications processes logical log of events, each event defines a change
// to current state of the object, i.e. which attributes were inserted,
// updated or deleted (a kind of diff). The event identifies the object that was
// changed together with using unique identifier.
//
// In contrast with other solutions, the event does not uses envelop approach.
// Instead, it side-car meta and data each other, making extendible
type Event[M, T any] struct {
	Meta *M `json:"meta,omitempty"`
	Data *T `json:"data,omitempty"`
}

// The default metadata associated with event.
type Meta struct {
	//
	// Unique identity for event.
	// It is automatically defined by the library upon the transmission unless
	// defined by sender. Preserving ID across sequence of messages allows
	// building request/response semantic.
	ID string `json:"id,omitempty"`

	//
	// Canonical IRI that defines a type of action.
	// It is automatically defined by the library upon the transmission unless
	// defined by sender.
	Type curie.IRI `json:"type,omitempty"`

	// Unique identity of the realm (logical environment or world) where the event was created.
	// Useful to support deployment isolation (e.g., green/blue, canary) in event-driven systems.
	Realm curie.IRI `json:"realm,omitempty"`

	//
	// Direct performer of the event, a software service that emits action to
	// the stream. It is automatically defined by the library upon the transmission
	// unless defined by sender.
	Agent curie.IRI `json:"agent,omitempty"`

	//
	// ISO8601 timestamps when action has been created
	// It is automatically defined by the library upon the transmission
	Created time.Time `json:"created"`

	//
	// Indicates target performer of the event, a software service that is able to
	Target curie.IRI `json:"target,omitempty"`

	//
	// Indirect participants, a user who initiated an event.
	Participant curie.IRI `json:"participant,omitempty"`
}
