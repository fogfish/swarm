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
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/emit"
)

type Event = swarm.Event[swarm.Meta, EventNote]

type EventNote struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func main() {
	q := eventbridge.Must(
		eventbridge.Emitter().
			WithKernel(swarm.WithAgent("eventbridge:example/event")).
			Build("swarm-example-eventbridge"),
	)

	bus := swarm.LogDeadLetters(emit.Event[Event](q))

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
