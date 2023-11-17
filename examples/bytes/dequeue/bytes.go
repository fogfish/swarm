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
	"os"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/bytes"
)

func main() {
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(os.Stdout,
				&slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			),
		),
	)

	q := queue.Must(sqs.New("swarm-test", swarm.WithLogStdErr()))

	go actor("user").handle(bytes.Dequeue(q, "User"))
	go actor("note").handle(bytes.Dequeue(q, "Note"))
	go actor("like").handle(bytes.Dequeue(q, "Like"))

	q.Await()
}

type actor string

func (a actor) handle(rcv <-chan *swarm.Msg[[]byte], ack chan<- *swarm.Msg[[]byte]) {
	for msg := range rcv {
		slog.Info("Event", "type", a, "msg", msg.Object)
		ack <- msg
	}
}
