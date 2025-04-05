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
	"github.com/fogfish/swarm/dequeue"
	"github.com/fogfish/swarm/kernel/encoding"
)

func main() {
	q, err := sqs.NewDequeuer("swarm-test",
		sqs.WithConfig(
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("sqs reader has failed", "err", err)
		return
	}

	go actor("user").handle(dequeue.Bytes(q, encoding.ForBytes("User")))
	go actor("note").handle(dequeue.Bytes(q, encoding.ForBytes("Note")))
	go actor("like").handle(dequeue.Bytes(q, encoding.ForBytes("Like")))

	q.Await()
}

type actor string

func (a actor) handle(rcv <-chan swarm.Msg[[]byte], ack chan<- swarm.Msg[[]byte]) {
	for msg := range rcv {
		slog.Info("Event", "type", a, "msg", msg.Object)
		ack <- msg
	}
}
