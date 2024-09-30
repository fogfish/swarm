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
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/enqueue"
)

type Event = swarm.Event[swarm.Meta, EventNote]

type EventNote struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func main() {
	q, err := eventbridge.NewEnqueuer("swarm-example-eventbridge",
		eventbridge.WithConfig(
			swarm.WithSource("swarm-example-eventbridge"),
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("eventbridge writer has failed", "err", err)
		return
	}

	bus := swarm.LogDeadLetters(enqueue.Event[swarm.Meta, EventNote](q))

	bus <- Event{
		Meta: &swarm.Meta{
			Type:        "note:EventCreateNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &EventNote{ID: "note", Text: "some text"},
	}

	bus <- Event{
		Meta: &swarm.Meta{
			Type:        "note:EventUpdateNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &EventNote{ID: "note", Text: "some text with changes"},
	}

	bus <- Event{
		Meta: &swarm.Meta{
			Type:        "note:EventRemoveNote",
			Agent:       "example",
			Participant: "user",
		},
		Data: &EventNote{ID: "note"},
	}

	q.Close()
}
