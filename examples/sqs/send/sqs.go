//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/sqs"
)

type NoteA struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type NoteB NoteA
type NoteC NoteA

func main() {
	sys := sqs.NewSystem("swarm-example-sqs")
	q := sqs.Must(sqs.New(sys, "swarm-test"))

	a, _ := queue.Send[*NoteA](q)
	b, _ := queue.Send[*NoteB](q)
	c, _ := queue.Send[*NoteC](q)

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	a <- &NoteA{ID: "a", Text: "message"}
	b <- &NoteB{ID: "b", Text: "message"}
	c <- &NoteC{ID: "c", Text: "message"}

	sys.Stop()
}
