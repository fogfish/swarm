//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/queue"
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
	q := queue.Must(eventbridge.New("swarm-example-eventbridge-latest",
		swarm.WithSource("swarm-example-eventbridge"),
		swarm.WithLogStdErr(),
	))

	user := queue.LogDeadLetters(queue.Enqueue[*User](q))
	note := queue.LogDeadLetters(queue.Enqueue[*Note](q))
	like := queue.LogDeadLetters(queue.Enqueue[*Like](q))

	user <- &User{ID: "user", Text: "some text"}
	note <- &Note{ID: "note", Text: "some text"}
	like <- &Like{ID: "like", Text: "some text"}

	q.Close()
}
