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
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/internal/qtest"
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
	qtest.NewLogger()

	// k, err := sqs.New("swarm-test")
	// // , &swarm.Config{
	// // 	Source:         "xxx",
	// // 	NetworkTimeout: 30 * time.Second,
	// // })
	// if err != nil {
	// 	panic(err)
	// }

	// go actor[User]("user").handle(
	// 	kernel.Dequeue[User](k.(*kernel.Kernel), "user", swarm.NewCodecJson[User]()),
	// 	// queue.Dequeue[User](q)
	// )

	// k.Await()

	q := queue.Must(sqs.New("swarm-test", swarm.WithLogStdErr()))

	go actor[User]("user").handle(queue.Dequeue[User](q))
	go actor[Note]("note").handle(queue.Dequeue[Note](q))
	go actor[Like]("like").handle(queue.Dequeue[Like](q))

	q.Await()
}

type actor[T any] string

func (a actor[T]) handle(rcv <-chan swarm.Msg[T], ack chan<- swarm.Msg[T]) {
	for msg := range rcv {
		slog.Info("Event", "type", a, "msg", msg.Object)
		ack <- msg
	}
}
