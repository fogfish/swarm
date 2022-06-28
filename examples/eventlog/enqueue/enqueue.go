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
	"github.com/fogfish/swarm/queue/events"
	"github.com/fogfish/swarm/queue/sqs"
)

//
// Date type (object) affected by events
type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

//
// Events
type EventCreateNote swarm.Event[*Note]

func (EventCreateNote) HKT1(swarm.EventType) {}
func (EventCreateNote) HKT2(*Note)           {}

type EventUpdateNote swarm.Event[*Note]

func (EventUpdateNote) HKT1(swarm.EventType) {}
func (EventUpdateNote) HKT2(*Note)           {}

type EventRemoveNote swarm.Event[*Note]

func (EventRemoveNote) HKT1(swarm.EventType) {}
func (EventRemoveNote) HKT2(*Note)           {}

func main() {
	sys := sqs.NewSystem("swarm-example-sqs")
	q := sqs.Must(sqs.New(sys, "swarm-test"))

	a, _ := events.Enqueue[*Note, EventCreateNote](q)
	b, _ := events.Enqueue[*Note, EventUpdateNote](q)
	c, _ := events.Enqueue[*Note, EventRemoveNote](q)
	d, _ := events.Enqueue[*Note, swarm.Event[*Note]](q)

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	//
	// Multiple channels emits events
	a <- &EventCreateNote{
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text"},
	}

	b <- &EventUpdateNote{
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text with changes"},
	}

	c <- &EventRemoveNote{
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note"},
	}

	//
	// Single channel emits event
	d <- &swarm.Event[*Note]{
		Type:        "note:EventCreateNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text"},
	}

	d <- &swarm.Event[*Note]{
		Type:        "note:EventUpdateNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text with changes"},
	}

	d <- &swarm.Event[*Note]{
		Type:        "note:EventRemoveNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note"},
	}

	sys.Close()
}
