//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"fmt"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventbridge"
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
	q, err := eventbridge.New("swarm-example-eventbridge")
	if err != nil {
		panic(err)
	}

	go actor[User]("a").handle(queue.Dequeue[User](q))
	go actor[Note]("b").handle(queue.Dequeue[Note](q))
	go actor[Like]("c").handle(queue.Dequeue[Like](q))

	q.Await()
}

//
type actor[T any] string

func (a actor[T]) handle(rcv <-chan *swarm.Msg[T], ack chan<- *swarm.Msg[T]) {
	for msg := range rcv {
		fmt.Printf("event on %s > %+v\n", a, msg.Object)
		ack <- msg
	}
}
