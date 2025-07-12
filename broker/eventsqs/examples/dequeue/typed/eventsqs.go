//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventsqs"
	"github.com/fogfish/swarm/dequeue"
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
	q := eventsqs.Channels().MustDequeuer()

	go actor[User]("user").handle(dequeue.Typed[User](q))
	go actor[Note]("note").handle(dequeue.Typed[Note](q))
	go actor[Like]("like").handle(dequeue.Typed[Like](q))

	q.Await()
}

type actor[T any] string

func (a actor[T]) handle(rcv <-chan swarm.Msg[T], ack chan<- swarm.Msg[T]) {
	for msg := range rcv {
		slog.Info("Event", "type", a, "msg", msg.Object)
		ack <- msg
	}
}
