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

type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Like struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

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

	user := swarm.LogDeadLetters(enqueue.Typed[*User](q))
	note := swarm.LogDeadLetters(enqueue.Typed[*Note](q))
	like := swarm.LogDeadLetters(enqueue.Typed[*Like](q))

	user <- &User{ID: "user", Text: "user signed  in"}
	note <- &Note{ID: "note", Text: "user wrote note"}
	like <- &Like{ID: "like", Text: "user liked note"}

	q.Close()
}
