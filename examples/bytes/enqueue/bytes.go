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

	user := queue.LogDeadLetters(bytes.Enqueue(q, "User"))
	note := queue.LogDeadLetters(bytes.Enqueue(q, "Note"))
	like := queue.LogDeadLetters(bytes.Enqueue(q, "Like"))

	user <- []byte("user|some text by user")

	note <- []byte("note|some note")

	like <- []byte("like|someone liked it")

	q.Close()
}
