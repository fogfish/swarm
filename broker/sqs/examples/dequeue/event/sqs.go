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
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/dequeue"
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
	q, err := sqs.NewDequeuer("swarm-test",
		sqs.WithConfig(
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("sqs reader has failed", "err", err)
		return
	}

	go create(dequeue.Event[swarm.Meta, CreatedUser](q))
	go update(dequeue.Event[swarm.Meta, UpdatedUser](q))
	go remove(dequeue.Event[swarm.Meta, RemovedUser](q))
	go common(dequeue.Event[swarm.Meta, Note](q))

	q.Await()
}

func create(rcv <-chan swarm.Evt[swarm.Meta, CreatedUser], ack chan<- swarm.Evt[swarm.Meta, CreatedUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "+ |", " ")
		fmt.Printf("create user > \n %s\n", v)
		ack <- msg
	}
}

func update(rcv <-chan swarm.Evt[swarm.Meta, UpdatedUser], ack chan<- swarm.Evt[swarm.Meta, UpdatedUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "~ |", " ")
		fmt.Printf("update user > \n %s\n", v)
		ack <- msg
	}
}

func remove(rcv <-chan swarm.Evt[swarm.Meta, RemovedUser], ack chan<- swarm.Evt[swarm.Meta, RemovedUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "- |", " ")
		fmt.Printf("remove user > \n %s\n", v)
		ack <- msg
	}
}

func common(rcv <-chan swarm.Evt[swarm.Meta, Note], ack chan<- swarm.Evt[swarm.Meta, Note]) {
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
		fmt.Printf("common note > \n %s\n", v)
		ack <- msg
	}
}
