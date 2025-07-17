//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"encoding/json"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// Base64 encoding for bytes, sent as JSON
type BytesCodec string

type packet struct {
	Octets []byte `json:"p,omitempty"`
}

func (c BytesCodec) Category() string { return string(c) }

func (c BytesCodec) Encode(obj []byte) (swarm.Bag, error) {
	msg, err := json.Marshal(packet{Octets: obj})
	if err != nil {
		return swarm.Bag{}, err
	}

	return swarm.Bag{
		Category: string(c),
		Object:   msg,
	}, nil
}
func (BytesCodec) Decode(bag swarm.Bag) ([]byte, error) {
	var pckt packet
	err := json.Unmarshal(bag.Object, &pckt)
	return pckt.Octets, err
}

// Creates bytes codec for Base64 encapsulated into Json "packet"
func ForBytes(cat string) BytesCodec { return BytesCodec(cat) }

func EmitBytes(q *kernel.EmitterIO, category string) (snd chan<- []byte, dlq <-chan []byte) {
	return kernel.EmitChan(q, ForBytes(category))
}

func RecvBytes(q *kernel.ListenerIO, category string) (<-chan swarm.Msg[[]byte], chan<- swarm.Msg[[]byte]) {
	return kernel.RecvChan(q, ForBytes(category))
}
