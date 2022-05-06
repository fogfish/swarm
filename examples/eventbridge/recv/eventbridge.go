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
)

func main() {
	sys := queue.System("test")
	queue := queue.Must(queue.EventBridge(sys, "swarm-test"))

	go actor("a").handle(queue.Recv("eventbridge.test.a"))
	go actor("b").handle(queue.Recv("eventbridge.test.b"))
	go actor("c").handle(queue.Recv("eventbridge.test.c"))

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	sys.Wait()
}

//
type actor string

func (a actor) handle(rcv <-chan swarm.Event, ack chan<- swarm.Event) {
	for msg := range rcv {
		logger.Debug("event on %s > %s", a, msg.Bytes())
		ack <- msg
	}
}
