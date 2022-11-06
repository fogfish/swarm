//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"time"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/sqsv2"
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
	q := sqsv2.New()

	// sys := sqs.NewSystem("swarm-example-sqs")
	// q := sqs.Must(sqs.New(sys, "swarm-test"))

	go actor[User]("a").handle(swarm.DequeueV2[User](q, "swarm-test-user"))
	// go actor[Note]("b").handle(queue.Dequeue[Note](q))
	// go actor[Like]("c").handle(queue.Dequeue[Like](q))

	// if err := sys.Listen(); err != nil {
	// 	panic(err)
	// }

	// sys.Wait()
	time.Sleep(1 * time.Second)
	q.Close()
}

//
type actor[T any] string

func (a actor[T]) handle(rcv <-chan *swarm.Msg[T], ack chan<- *swarm.Msg[T]) {
	for msg := range rcv {
		logger.Debug("event on %s > %+v", a, msg.Object)
		ack <- msg
	}
}
