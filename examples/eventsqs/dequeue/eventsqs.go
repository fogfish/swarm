//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventsqs"
	"github.com/fogfish/swarm/queue"
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
	q, err := eventsqs.New("swarm-example-sqs-latest")
	if err != nil {
		panic(err)
	}

	go actor[User]("user").handle(queue.Dequeue[User](q))
	go actor[Note]("note").handle(queue.Dequeue[Note](q))
	go actor[Like]("like").handle(queue.Dequeue[Like](q))

	q.Await()
}

//
type actor[T any] string

func (a actor[T]) handle(rcv <-chan *swarm.Msg[T], ack chan<- *swarm.Msg[T]) {
	for msg := range rcv {
		logger.Debug("event on %s > %+v", a, msg.Object)
		ack <- msg
	}
}