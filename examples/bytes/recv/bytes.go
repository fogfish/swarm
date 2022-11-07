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
	"github.com/fogfish/swarm/broker/sqs"
	queue "github.com/fogfish/swarm/queue/bytes"
)

func main() {
	q, err := sqs.New("swarm-test")
	if err != nil {
		panic(err)
	}

	go actor("user").handle(queue.Dequeue(q, "User"))
	go actor("note").handle(queue.Dequeue(q, "Note"))
	go actor("like").handle(queue.Dequeue(q, "Like"))

	q.Await()
}

//
type actor string

func (a actor) handle(rcv <-chan *swarm.Msg[[]byte], ack chan<- *swarm.Msg[[]byte]) {
	for msg := range rcv {
		logger.Debug("event on %s > %s", a, msg.Object)
		ack <- msg
	}
}
