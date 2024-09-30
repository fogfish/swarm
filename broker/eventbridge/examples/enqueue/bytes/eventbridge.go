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
	"github.com/fogfish/swarm/enqueue"
)

func main() {
	q, err := eventbridge.NewEnqueuer("swarm-example-eventbridge",
		eventbridge.WithConfig(
			swarm.WithSource("swarm-example-eventbridge"),
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("eventbridge writer has failed", "err", err)
		return
	}

	user := swarm.LogDeadLetters(enqueue.Bytes(q, "User"))
	note := swarm.LogDeadLetters(enqueue.Bytes(q, "Note"))
	like := swarm.LogDeadLetters(enqueue.Bytes(q, "Like"))

	user <- []byte(`User Signed in`)
	note <- []byte(`User wrote note`)
	like <- []byte(`User liked note`)

	q.Close()
}
