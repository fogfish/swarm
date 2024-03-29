//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"encoding/json"
	"fmt"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/internal/qtest"
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

// Events
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
	qtest.NewLogger()

	q := queue.Must(sqs.New("swarm-test", swarm.WithLogStdErr()))

	go create(events.Dequeue[*User, EventCreateUser](q))
	go update(events.Dequeue[*User, EventUpdateUser](q))
	go remove(events.Dequeue[*User, EventRemoveUser](q))
	go common(events.Dequeue[*Note, EventNote](q))

	q.Await()
}

func create(rcv <-chan swarm.Msg[*EventCreateUser], ack chan<- swarm.Msg[*EventCreateUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "+ |", " ")
		fmt.Printf("create user > \n %s\n", v)
		ack <- msg
	}
}

func update(rcv <-chan swarm.Msg[*EventUpdateUser], ack chan<- swarm.Msg[*EventUpdateUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "~ |", " ")
		fmt.Printf("update user > \n %s\n", v)
		ack <- msg
	}
}

func remove(rcv <-chan swarm.Msg[*EventRemoveUser], ack chan<- swarm.Msg[*EventRemoveUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "- |", " ")
		fmt.Printf("remove user > \n %s\n", v)
		ack <- msg
	}
}

func common(rcv <-chan swarm.Msg[*EventNote], ack chan<- swarm.Msg[*EventNote]) {
	for msg := range rcv {
		prefix := ""
		switch string(msg.Object.Type) {
		case "note:EventCreateNote":
			prefix = "+ |"
		case "note:EventUpdateNote":
			prefix = "~ |"
		case "note:EventRemoveNote":
			prefix = "- |"
		}

		v, _ := json.MarshalIndent(msg, prefix, " ")
		fmt.Printf("common note > \n %s\n", v)
		ack <- msg
	}
}
