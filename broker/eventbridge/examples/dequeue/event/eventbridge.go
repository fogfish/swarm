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
	"github.com/fogfish/swarm/dequeue"
)

type Event = swarm.Event[swarm.Meta, EventNote]

type EventNote struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func main() {
	q, err := eventbridge.NewDequeuer(
		eventbridge.WithConfig(
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("eventbridge reader has failed", "err", err)
		return
	}

	go bus(dequeue.Event[swarm.Meta, EventNote](q))

	q.Await()
}

func bus(rcv <-chan swarm.Msg[Event], ack chan<- swarm.Msg[Event]) {
	for msg := range rcv {
		prefix := ""
		switch string(msg.Object.Meta.Type) {
		case "note:EventCreateNote":
			prefix = "+ |"
		case "note:EventUpdateNote":
			prefix = "~ |"
		case "note:EventRemoveNote":
			prefix = "- |"
		}

		v, _ := json.MarshalIndent(msg, prefix, " ")
		fmt.Printf("event > \n %s\n", v)

		ack <- msg
	}
}
