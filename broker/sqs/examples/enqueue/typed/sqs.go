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
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/emit"
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
	q := sqs.Must(sqs.Emitter().Build("swarm-test"))

	user := swarm.LogDeadLetters(emit.Typed[*User](q))
	note := swarm.LogDeadLetters(emit.Typed[*Note](q))
	like := swarm.LogDeadLetters(emit.Typed[*Like](q))

	user <- &User{ID: "user", Text: "some text by user"}

	note <- &Note{ID: "note", Text: "some note"}

	like <- &Like{ID: "like", Text: "someone liked it"}

	q.Close()
}
