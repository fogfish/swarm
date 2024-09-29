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
	q, err := sqs.NewEnqueuer("swarm-test",
		sqs.WithConfig(
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("sqs writer has failed", "err", err)
		return
	}

	user := swarm.LogDeadLetters(enqueue.Typed[*User](q))
	note := swarm.LogDeadLetters(enqueue.Typed[*Note](q))
	like := swarm.LogDeadLetters(enqueue.Typed[*Like](q))

	user <- &User{ID: "user", Text: "some text by user"}

	note <- &Note{ID: "note", Text: "some note"}

	like <- &Like{ID: "like", Text: "someone liked it"}

	q.Close()
}
