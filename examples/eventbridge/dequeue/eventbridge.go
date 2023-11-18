//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

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

	q := queue.Must(eventbridge.New("swarm-example-eventbridge", swarm.WithLogStdErr()))

	go actor[User]("user").handle(queue.Dequeue[User](q))
	go actor[Note]("note").handle(queue.Dequeue[Note](q))
	go actor[Like]("like").handle(queue.Dequeue[Like](q))
	go ebus(events.Dequeue[*Note, EventNote](q))

	q.Await()
}

type actor[T any] string

func (a actor[T]) handle(rcv <-chan *swarm.Msg[T], ack chan<- *swarm.Msg[T]) {
	for msg := range rcv {
		slog.Info("Event", "type", a, "msg", msg.Object)
		ack <- msg
	}
}

func ebus(rcv <-chan *EventNote, ack chan<- *EventNote) {
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
		fmt.Printf("common note > \n %s\n", v)
		ack <- msg
	}
}
