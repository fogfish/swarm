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

// Json codec for I/O kernel
type CodecJson[T any] string

func (c CodecJson[T]) Category() string { return string(c) }

func (CodecJson[T]) Encode(x T) ([]byte, error) {
	return json.Marshal(x)
}

func (CodecJson[T]) Decode(b []byte) (x T, err error) {
	err = json.Unmarshal(b, &x)
	return
}

func NewCodecJson[T any](category ...string) CodecJson[T] {
	return CodecJson[T](swarm.TypeOf[T](category...))
}

//------------------------------------------------------------------------------

// Byte identity codec for I/O kernet
type CodecByte string

func (c CodecByte) Category() string              { return string(c) }
func (CodecByte) Encode(x []byte) ([]byte, error) { return x, nil }
func (CodecByte) Decode(x []byte) ([]byte, error) { return x, nil }

func NewCodecByte(cat string) CodecByte { return CodecByte(cat) }

//------------------------------------------------------------------------------

// Encode Bytes as "JSON packet"
type CodecPacket string

type packet struct {
	Octets []byte `json:"p,omitempty"`
}

func (c CodecPacket) Category() string { return string(c) }

func (CodecPacket) Encode(x []byte) ([]byte, error) {
	b, err := json.Marshal(packet{Octets: x})
	return b, err
}
func (CodecPacket) Decode(x []byte) ([]byte, error) {
	var pckt packet
	err := json.Unmarshal(x, &pckt)
	return pckt.Octets, err
}

func NewCodecPacket(cat string) CodecPacket { return CodecPacket(cat) }

//------------------------------------------------------------------------------

// Event codec for I/O kernel
type CodecEvent[M, T any] struct {
	source string
	cat    string
	shape  optics.Lens4[M, string, curie.IRI, curie.IRI, time.Time]
}

func (c CodecEvent[M, T]) Category() string { return c.cat }

func (c CodecEvent[M, T]) Encode(obj swarm.Event[M, T]) ([]byte, error) {
	_, knd, src, _ := c.shape.Get(obj.Meta)
	if knd == "" {
		knd = curie.IRI(c.cat)
	}

	if src == "" {
		src = curie.IRI(c.source)
	}

	c.shape.Put(obj.Meta, guid.G(guid.Clock).String(), knd, src, time.Now())

	return json.Marshal(obj)
}

func (c CodecEvent[M, T]) Decode(b []byte) (swarm.Event[M, T], error) {
	var x swarm.Event[M, T]
	err := json.Unmarshal(b, &x)

	return x, err
}

func NewCodecEvent[M, T any](source string, category ...string) CodecEvent[M, T] {
	return CodecEvent[M, T]{
		source: source,
		cat:    swarm.TypeOf[T](category...),
		shape:  optics.ForShape4[M, string, curie.IRI, curie.IRI, time.Time]("ID", "Type", "Agent", "Created"),
	}
}
