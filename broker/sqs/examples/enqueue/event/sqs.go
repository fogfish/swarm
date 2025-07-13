//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/emit"
)

// Date type (object) affected by events
type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type CreatedUser User

type EvtCreatedUser = swarm.Event[swarm.Meta, CreatedUser]

type UpdatedUser User

type EvtUpdatedUser = swarm.Event[swarm.Meta, UpdatedUser]

type RemovedUser User

type EvtRemovedUser = swarm.Event[swarm.Meta, RemovedUser]

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type EvtNote = swarm.Event[swarm.Meta, Note]

func main() {
	q := sqs.Must(sqs.Emitter().Build("swarm-test"))

	userCreated := swarm.LogDeadLetters(emit.Event[EvtCreatedUser](q))
	userUpdated := swarm.LogDeadLetters(emit.Event[EvtUpdatedUser](q))
	userRemoved := swarm.LogDeadLetters(emit.Event[EvtRemovedUser](q))
	note := swarm.LogDeadLetters(emit.Event[EvtNote](q))

	//
	// Multiple channels emits events
	userCreated <- EvtCreatedUser{
		Meta: &swarm.Meta{
			Agent:       "example",
			Participant: "user",
		},
		Data: &CreatedUser{ID: "user", Text: "some text"},
	}

	userUpdated <- EvtUpdatedUser{
		Meta: &swarm.Meta{
			Agent:       "example",
			Participant: "user",
		},
		Data: &UpdatedUser{ID: "user", Text: "some text with changes"},
	}

	userRemoved <- EvtRemovedUser{
		Meta: &swarm.Meta{
			Agent:       "example",
			Participant: "user",
		},
		Data: &RemovedUser{ID: "user"},
	}

	//
	// Single channel emits event
	note <- EvtNote{
		Meta: &swarm.Meta{
			Type:        "note:EventCreateNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &Note{ID: "note", Text: "some text"},
	}

	note <- EvtNote{
		Meta: &swarm.Meta{
			Type:        "note:EventUpdateNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &Note{ID: "note", Text: "some text with changes"},
	}

	note <- EvtNote{
		Meta: &swarm.Meta{
			Type:        "note:EventRemoveNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &Note{ID: "note"},
	}

	q.Close()
}
