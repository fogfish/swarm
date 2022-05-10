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
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/sqs"
)

type NoteA struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type NoteB NoteA
type NoteC NoteA

func main() {
	sys := sqs.NewSystem("swarm-example-sqs")
	q := sqs.Must(sqs.New(sys, "swarm-test"))

	go actor[NoteA]("a").handle(queue.Recv[NoteA](q))
	go actor[NoteB]("b").handle(queue.Recv[NoteB](q))
	go actor[NoteC]("c").handle(queue.Recv[NoteC](q))

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	sys.Wait()
}

//
type actor[T any] string

func (a actor[T]) handle(rcv <-chan *swarm.Msg[T], ack chan<- *swarm.Msg[T]) {
	for msg := range rcv {
		logger.Debug("event on %s > %+v", a, msg.Object)
		ack <- msg
	}
}
