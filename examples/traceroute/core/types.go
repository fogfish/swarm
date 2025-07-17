//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package core

import (
	"github.com/fogfish/curie/v2"
	"github.com/fogfish/guid/v2"
	"github.com/fogfish/swarm"
)

type ReqTrace = swarm.Event[swarm.Meta, Request]

type Request struct {
	// WebSocket channel identifier
	Channel string `json:"channel,omitempty"`

	// Opaque data attached to each trace and response
	Opaque string `json:"opaque,omitempty"`
}

func ToRequest(evt ReqTrace) Request {
	if evt.Data == nil {
		return Request{
			Channel: string(evt.Digest),
		}
	}

	req := Request{
		Channel: evt.Data.Channel,
		Opaque:  evt.Data.Opaque,
	}

	if req.Channel == "" {
		req.Channel = string(evt.Digest)
	}

	return req
}

func ToReqTrace(req Request) ReqTrace {
	return ReqTrace{
		Meta: &swarm.Meta{
			Sink: curie.IRI(req.Channel),
		},
		Data: &req,
	}
}

type ActStatus = swarm.Event[swarm.Meta, Status]

type Status struct {
	// Host name
	Host string `json:"host,omitempty"`

	// Opaque data attached to each trace and response
	Opaque string `json:"opaque,omitempty"`
}

type RequestDB struct {
	HKey    curie.IRI `dynamodbav:"prefix,omitempty" json:"prefix,omitempty"`
	SKey    curie.IRI `dynamodbav:"suffix,omitempty" json:"suffix,omitempty"`
	Channel string    `dynamodbav:"channel,omitempty" json:"channel,omitempty"`
	Opaque  string    `dynamodbav:"opaque,omitempty" json:"opaque,omitempty"`
}

func (req RequestDB) HashKey() curie.IRI { return req.HKey }
func (req RequestDB) SortKey() curie.IRI { return req.SKey }

func ToRequestDB(evt ReqTrace) RequestDB {
	if evt.Data == nil {
		return RequestDB{
			HKey:    curie.IRI(guid.G(guid.Clock).String()),
			SKey:    curie.IRI(guid.G(guid.Clock).String()),
			Channel: string(evt.Digest),
		}
	}

	req := RequestDB{
		HKey:    curie.IRI(guid.G(guid.Clock).String()),
		SKey:    curie.IRI(guid.G(guid.Clock).String()),
		Channel: evt.Data.Channel,
		Opaque:  evt.Data.Opaque,
	}

	if req.Channel == "" {
		req.Channel = string(evt.Digest)
	}

	return req
}

func FromRequestDB(val RequestDB) ReqTrace {
	return ReqTrace{
		Meta: &swarm.Meta{
			Sink: curie.IRI(val.Channel),
		},
		Data: &Request{
			Channel: val.Channel,
			Opaque:  val.Opaque,
		},
	}
}
