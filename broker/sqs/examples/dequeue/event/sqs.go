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
	dequeue "github.com/fogfish/swarm/listen"
)

// Date type (object) affected by events
type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type CreatedUser User

type EvtCreatedUser = swarm.Event[swarm.Meta, CreatedUser]

type UpdatedUser User

type EvtUpdatedUser = swarm.Event[swarm.Meta, UpdatedUser]

type RemovedUser User

type EvtRemovedUser = swarm.Event[swarm.Meta, RemovedUser]

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type EvtNote = swarm.Event[swarm.Meta, Note]

func main() {
	q := sqs.Must(sqs.Listener().Build("swarm-test"))

	go create(dequeue.Event[EvtCreatedUser](q))
	go update(dequeue.Event[EvtUpdatedUser](q))
	go remove(dequeue.Event[EvtRemovedUser](q))
	go common(dequeue.Event[EvtNote](q))

	q.Await()
}

func create(rcv <-chan swarm.Msg[EvtCreatedUser], ack chan<- swarm.Msg[EvtCreatedUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg.Object, "+ |", " ")
		fmt.Printf("create user > \n %s\n", v)
		ack <- msg
	}
}

func update(rcv <-chan swarm.Msg[EvtUpdatedUser], ack chan<- swarm.Msg[EvtUpdatedUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg.Object, "~ |", " ")
		fmt.Printf("update user > \n %s\n", v)
		ack <- msg
	}
}

func remove(rcv <-chan swarm.Msg[EvtRemovedUser], ack chan<- swarm.Msg[EvtRemovedUser]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg.Object, "- |", " ")
		fmt.Printf("remove user > \n %s\n", v)
		ack <- msg
	}
}

func common(rcv <-chan swarm.Msg[EvtNote], ack chan<- swarm.Msg[EvtNote]) {
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

		v, _ := json.MarshalIndent(msg.Object, prefix, " ")
		fmt.Printf("common note > \n %s\n", v)
		ack <- msg
	}
}
