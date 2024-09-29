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
	"github.com/fogfish/swarm/enqueue"
)

func main() {
	q, err := sqs.NewEnqueuer("swarm-test",
		sqs.WithConfig(
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("sqs writer has failed", "err", err)
		return
	}

	user := swarm.LogDeadLetters(enqueue.Bytes(q, "User"))
	note := swarm.LogDeadLetters(enqueue.Bytes(q, "Note"))
	like := swarm.LogDeadLetters(enqueue.Bytes(q, "Like"))

	user <- []byte("user|some text by user")

	note <- []byte("note|some note")

	like <- []byte("like|someone liked it")

	q.Close()
}
