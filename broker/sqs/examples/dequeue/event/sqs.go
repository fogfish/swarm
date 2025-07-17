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
	"github.com/fogfish/swarm/listen"
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

	go create(listen.Event[EvtCreatedUser](q))
	go update(listen.Event[EvtUpdatedUser](q))
	go remove(listen.Event[EvtRemovedUser](q))
	go common(listen.Event[EvtNote](q))

	q.Await()
}

func create(rcv <-chan EvtCreatedUser, ack chan<- EvtCreatedUser) {
	for evt := range rcv {
		v, _ := json.MarshalIndent(evt, "+ |", " ")
		fmt.Printf("create user > \n %s\n", v)
		ack <- evt
	}
}

func update(rcv <-chan EvtUpdatedUser, ack chan<- EvtUpdatedUser) {
	for evt := range rcv {
		v, _ := json.MarshalIndent(evt, "~ |", " ")
		fmt.Printf("update user > \n %s\n", v)
		ack <- evt
	}
}

func remove(rcv <-chan EvtRemovedUser, ack chan<- EvtRemovedUser) {
	for evt := range rcv {
		v, _ := json.MarshalIndent(evt, "- |", " ")
		fmt.Printf("remove user > \n %s\n", v)
		ack <- evt
	}
}

func common(rcv <-chan EvtNote, ack chan<- EvtNote) {
	for msg := range rcv {
		prefix := ""
		switch string(msg.Meta.Type) {
		case "note:EventCreateNote":
			prefix = "+ |"
		case "note:EventUpdateNote":
			prefix = "~ |"
		case "note:EventRemoveNote":
			prefix = "- |"
		}

		v, _ := json.MarshalIndent(msg.Data, prefix, " ")
		fmt.Printf("common note > \n %s\n", v)
		ack <- msg
	}
}
