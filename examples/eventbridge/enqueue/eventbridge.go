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
	"github.com/fogfish/swarm/internal/qtest"
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/events"
)

type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Like struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type EventNote swarm.Event[*Note]

func (EventNote) HKT1(swarm.EventType) {}
func (EventNote) HKT2(*Note)           {}

func main() {
	qtest.NewLogger()

	q := queue.Must(eventbridge.New("swarm-example-eventbridge-latest",
		swarm.WithSource("swarm-example-eventbridge"),
		swarm.WithLogStdErr(),
	))

	user := swarm.LogDeadLetters(queue.Enqueue[*User](q))
	note := swarm.LogDeadLetters(queue.Enqueue[*Note](q))
	like := swarm.LogDeadLetters(queue.Enqueue[*Like](q))
	ebus := swarm.LogDeadLetters(events.Enqueue[*Note, EventNote](q))

	user <- &User{ID: "user", Text: "some text"}
	note <- &Note{ID: "note", Text: "some text"}
	like <- &Like{ID: "like", Text: "some text"}

	//
	// Single channel emits event
	ebus <- &EventNote{
		Type:        "note:EventCreateNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text"},
	}

	ebus <- &EventNote{
		Type:        "note:EventUpdateNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text with changes"},
	}

	ebus <- &EventNote{
		Type:        "note:EventRemoveNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note"},
	}

	q.Close()
}
