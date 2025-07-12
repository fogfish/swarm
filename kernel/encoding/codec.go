//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package encoding

import (
	"encoding/json"
	"time"

	"github.com/fogfish/curie/v2"
	"github.com/fogfish/golem/optics"
	"github.com/fogfish/guid/v2"
	"github.com/fogfish/swarm"
)

//------------------------------------------------------------------------------

// JSON encoding for typed objects
type Typed[T any] string

func (c Typed[T]) Category() string { return string(c) }

func (Typed[T]) Encode(x T) ([]byte, error) {
	return json.Marshal(x)
}

func (Typed[T]) Decode(b []byte) (x T, err error) {
	err = json.Unmarshal(b, &x)
	return
}

// Create JSON codec for typed objects
func ForTyped[T any](category ...string) Typed[T] {
	return Typed[T](swarm.TypeOf[T](category...))
}

//------------------------------------------------------------------------------

// JSON encoding for events
type Event[M, T any] struct {
	realm curie.IRI
	agent curie.IRI
	cat   curie.IRI
	shape optics.Lens5[M, string, curie.IRI, curie.IRI, curie.IRI, time.Time]
}

func (c Event[M, T]) Category() string { return string(c.cat) }

func (c Event[M, T]) Encode(obj swarm.Event[M, T]) ([]byte, error) {
	_, cat, rlm, agt, _ := c.shape.Get(obj.Meta)
	if cat == "" {
		cat = c.cat
	}

	if agt == "" {
		agt = c.agent
	}

	if rlm == "" {
		rlm = c.realm
	}

	c.shape.Put(obj.Meta, guid.G(guid.Clock).String(), cat, rlm, agt, time.Now())

	return json.Marshal(obj)
}

func (c Event[M, T]) Decode(b []byte) (swarm.Event[M, T], error) {
	var x swarm.Event[M, T]
	err := json.Unmarshal(b, &x)

	return x, err
}

// Creates JSON codec for events
func ForEvent[E swarm.Event[M, T], M, T any](realm, agent string, category ...string) Event[M, T] {
	return Event[M, T]{
		realm: curie.IRI(realm),
		agent: curie.IRI(agent),
		cat:   curie.IRI(swarm.TypeOf[T](category...)),
		shape: optics.ForShape5[M, string, curie.IRI, curie.IRI, curie.IRI, time.Time]("ID", "Type", "Realm", "Agent", "Created"),
	}
}

//------------------------------------------------------------------------------

// Identity encoding for bytes
type Bytes string

func (c Bytes) Category() string              { return string(c) }
func (Bytes) Encode(x []byte) ([]byte, error) { return x, nil }
func (Bytes) Decode(x []byte) ([]byte, error) { return x, nil }

// Create bytes identity codec
func ForBytes(cat string) Bytes { return Bytes(cat) }

//------------------------------------------------------------------------------

// Base64 encoding for bytes, sent as JSON
type BytesJB64 string

type packet struct {
	Octets []byte `json:"p,omitempty"`
}

func (c BytesJB64) Category() string { return string(c) }

func (BytesJB64) Encode(x []byte) ([]byte, error) {
	b, err := json.Marshal(packet{Octets: x})
	return b, err
}
func (BytesJB64) Decode(x []byte) ([]byte, error) {
	var pckt packet
	err := json.Unmarshal(x, &pckt)
	return pckt.Octets, err
}

// Creates bytes codec for Base64 encapsulated into Json "packat"
func ForBytesJB64(cat string) BytesJB64 { return BytesJB64(cat) }
