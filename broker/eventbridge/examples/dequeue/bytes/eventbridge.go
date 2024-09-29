//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"log/slog"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/dequeue"
)

func main() {
	q, err := eventbridge.NewReader("swarm-example-eventbridge",
		eventbridge.WithConfig(
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("eventbridge reader has failed", "err", err)
		return
	}

	//
	go actor("user").handle(dequeue.Bytes(q, "User"))
	go actor("note").handle(dequeue.Bytes(q, "Note"))
	go actor("like").handle(dequeue.Bytes(q, "Like"))

	q.Await()
}

type actor string

func (a actor) handle(rcv <-chan swarm.Msg[[]byte], ack chan<- swarm.Msg[[]byte]) {
	for msg := range rcv {
		slog.Info("Event", "type", a, "msg", string(msg.Object))
		ack <- msg
	}
}