//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/emit"
	"github.com/fogfish/swarm/listen"
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

func main() {
	// q := eventbridge.Channels().MustDequeuer()
	q := eventbridge.Must(eventbridge.Endpoint().Build("swarm-example-eventbridge"))

	note := swarm.LogDeadLetters(emit.Typed[Note](q.EmitterCore))
	like := swarm.LogDeadLetters(emit.Typed[Like](q.EmitterCore))

	go (&xxx{note, like}).handle(listen.Typed[User](q.ListenerCore))

	//
	// go actor[User]("user").handle(dequeue.Typed[User](q.Dequeuer))
	go actor[Note]("note").handle(listen.Typed[Note](q.ListenerCore))
	go actor[Like]("like").handle(listen.Typed[Like](q.ListenerCore))

	q.Await()
}

type xxx struct {
	note chan<- Note
	like chan<- Like
}

func (x *xxx) handle(rcv <-chan swarm.Msg[User], ack chan<- swarm.Msg[User]) {
	for msg := range rcv {
		slog.Info("Event", "type", "user", "msg", msg.Object)
		x.note <- Note{ID: "note", Text: "user wrote note"}
		x.like <- Like{ID: "like", Text: "user liked note"}
		ack <- msg
	}
}

type actor[T any] string

func (a actor[T]) handle(rcv <-chan swarm.Msg[T], ack chan<- swarm.Msg[T]) {
	for msg := range rcv {
		slog.Info("Event", "type", a, "msg", msg.Object)
		ack <- msg
	}
}
