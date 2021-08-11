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
	"github.com/fogfish/swarm/queue/sqs"
)

func main() {
	sys := swarm.New("test")
	queue := swarm.Must(sqs.New(sys, "swarm-test"))

	go actor("a").handle(queue.Recv("sqs.test.a"))
	go actor("b").handle(queue.Recv("sqs.test.b"))
	go actor("c").handle(queue.Recv("sqs.test.c"))

	sys.Wait()
}

//
type actor string

func (a actor) handle(rcv <-chan swarm.Msg, ack chan<- swarm.Msg) {
	for msg := range rcv {
		logger.Debug("event on %s > %s", a, msg.Bytes())
		ack <- msg
	}
}
