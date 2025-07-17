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

func (c Typed[T]) Encode(obj T) (swarm.Bag, error) {
	msg, err := json.Marshal(obj)
	if err != nil {
		return swarm.Bag{}, err
	}

	return swarm.Bag{
		Category: string(c),
		Object:   msg,
	}, nil
}

func (Typed[T]) Decode(bag swarm.Bag) (x T, err error) {
	err = json.Unmarshal(bag.Object, &x)
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

func (c Event[M, T]) Encode(obj swarm.Event[M, T]) (swarm.Bag, error) {
	if obj.Meta == nil {
		obj.Meta = new(M)
	}

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
	msg, err := json.Marshal(obj)
	if err != nil {
		return swarm.Bag{}, err
	}

	return swarm.Bag{
		Category: string(cat),
		Object:   msg,
	}, nil
}

func (c Event[M, T]) Decode(bag swarm.Bag) (swarm.Event[M, T], error) {
	var x swarm.Event[M, T]
	err := json.Unmarshal(bag.Object, &x)

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

func (c Bytes) Category() string { return string(c) }

func (c Bytes) Encode(x []byte) (swarm.Bag, error) {
	return swarm.Bag{
		Category: string(c),
		Object:   x,
	}, nil
}

func (Bytes) Decode(bag swarm.Bag) ([]byte, error) {
	return bag.Object, nil
}

// Create bytes identity codec
func ForBytes(cat string) Bytes { return Bytes(cat) }
