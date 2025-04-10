//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/enqueue"
)

// Date type (object) affected by events
type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type CreatedUser User

type UpdatedUser User

type RemovedUser User

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func main() {
	q, err := sqs.NewEnqueuer("swarm-test",
		sqs.WithConfig(
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("sqs writer has failed", "err", err)
		return
	}

	userCreated := swarm.LogDeadLetters(enqueue.Event[swarm.Meta, CreatedUser](q))
	userUpdated := swarm.LogDeadLetters(enqueue.Event[swarm.Meta, UpdatedUser](q))
	userRemoved := swarm.LogDeadLetters(enqueue.Event[swarm.Meta, RemovedUser](q))
	note := swarm.LogDeadLetters(enqueue.Event[swarm.Meta, Note](q))

	//
	// Multiple channels emits events
	userCreated <- swarm.Event[swarm.Meta, CreatedUser]{
		Meta: &swarm.Meta{
			Agent:       "example",
			Participant: "user",
		},
		Data: &CreatedUser{ID: "user", Text: "some text"},
	}

	userUpdated <- swarm.Event[swarm.Meta, UpdatedUser]{
		Meta: &swarm.Meta{
			Agent:       "example",
			Participant: "user",
		},
		Data: &UpdatedUser{ID: "user", Text: "some text with changes"},
	}

	userRemoved <- swarm.Event[swarm.Meta, RemovedUser]{
		Meta: &swarm.Meta{
			Agent:       "example",
			Participant: "user",
		},
		Data: &RemovedUser{ID: "user"},
	}

	//
	// Single channel emits event
	note <- swarm.Event[swarm.Meta, Note]{
		Meta: &swarm.Meta{
			Type:        "note:EventCreateNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &Note{ID: "note", Text: "some text"},
	}

	note <- swarm.Event[swarm.Meta, Note]{
		Meta: &swarm.Meta{
			Type:        "note:EventUpdateNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &Note{ID: "note", Text: "some text with changes"},
	}

	note <- swarm.Event[swarm.Meta, Note]{
		Meta: &swarm.Meta{
			Type:        "note:EventRemoveNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &Note{ID: "note"},
	}

	q.Close()
}
