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
	"github.com/fogfish/swarm/queue/eventsqs"
)

func main() {
	sys := eventsqs.NewSystem("swarm-example-eventsqs")
	q := eventsqs.Must(eventsqs.New(sys, "swarm-test"))

	go actor("a").handle(q.Recv("sqs.test.a"))
	go actor("b").handle(q.Recv("sqs.test.b"))
	go actor("c").handle(q.Recv("sqs.test.c"))

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	sys.Wait()
}

//
type actor string

func (a actor) handle(rcv <-chan swarm.Object, ack chan<- swarm.Object) {
	for msg := range rcv {
		logger.Debug("event on %s > %s", a, msg.Bytes())
		ack <- msg
	}
}
