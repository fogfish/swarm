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
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/events"
)

// Date type (object) affected by events
type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type EventCreateUser swarm.Event[*User]

func (EventCreateUser) HKT1(swarm.EventType) {}
func (EventCreateUser) HKT2(*User)           {}

type EventUpdateUser swarm.Event[*User]

func (EventUpdateUser) HKT1(swarm.EventType) {}
func (EventUpdateUser) HKT2(*User)           {}

type EventRemoveUser swarm.Event[*User]

func (EventRemoveUser) HKT1(swarm.EventType) {}
func (EventRemoveUser) HKT2(*User)           {}

type EventNote swarm.Event[*Note]

func (EventNote) HKT1(swarm.EventType) {}
func (EventNote) HKT2(*Note)           {}

func main() {
	q := queue.Must(sqs.New("swarm-test", swarm.WithLogStdErr()))

	userCreated := queue.LogDeadLetters(events.Enqueue[*User, EventCreateUser](q))
	userUpdated := queue.LogDeadLetters(events.Enqueue[*User, EventUpdateUser](q))
	userRemoved := queue.LogDeadLetters(events.Enqueue[*User, EventRemoveUser](q))
	note := queue.LogDeadLetters(events.Enqueue[*Note, EventNote](q))

	//
	// Multiple channels emits events
	userCreated <- &EventCreateUser{
		Agent:       "example",
		Participant: "user",
		Object:      &User{ID: "user", Text: "some text"},
	}

	userUpdated <- &EventUpdateUser{
		Agent:       "example",
		Participant: "user",
		Object:      &User{ID: "user", Text: "some text with changes"},
	}

	userRemoved <- &EventRemoveUser{
		Agent:       "example",
		Participant: "user",
		Object:      &User{ID: "note"},
	}

	//
	// Single channel emits event
	note <- &EventNote{
		Type:        "note:EventCreateNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text"},
	}

	note <- &EventNote{
		Type:        "note:EventUpdateNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text with changes"},
	}

	note <- &EventNote{
		Type:        "note:EventRemoveNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note"},
	}

	q.Close()
}
