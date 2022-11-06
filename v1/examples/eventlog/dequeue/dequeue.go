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

	"github.com/fogfish/logger"
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

	go create(events.Dequeue[*Note, EventCreateNote](q))
	go update(events.Dequeue[*Note, EventUpdateNote](q))
	go remove(events.Dequeue[*Note, EventRemoveNote](q))
	go common(events.Dequeue[*Note, swarm.Event[*Note]](q))

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	sys.Wait()
}

func create(rcv <-chan *EventCreateNote, ack chan<- *EventCreateNote) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "+ |", " ")
		logger.Debug("create note > \n %s\n", v)
		ack <- msg
	}
}

func update(rcv <-chan *EventUpdateNote, ack chan<- *EventUpdateNote) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "~ |", " ")
		logger.Debug("update note > \n %s\n", v)
		ack <- msg
	}
}

func remove(rcv <-chan *EventRemoveNote, ack chan<- *EventRemoveNote) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "- |", " ")
		logger.Debug("remove note > \n %s\n", v)
		ack <- msg
	}
}

func common(rcv <-chan *swarm.Event[*Note], ack chan<- *swarm.Event[*Note]) {
	for msg := range rcv {
		prefix := ""
		switch string(msg.Type) {
		case "note:EventCreateNote":
			prefix = "+ |"
		case "note:EventUpdateNote":
			prefix = "~ |"
		case "note:EventRemoveNote":
			prefix = "- |"
		}

		v, _ := json.MarshalIndent(msg, prefix, " ")
		logger.Debug("common note > \n %s\n", v)
		ack <- msg
	}
}
