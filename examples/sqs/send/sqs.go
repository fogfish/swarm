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
	sys := sqs.NewSystem("swarm-example-sqs")
	q := sqs.Must(sqs.New(sys, "swarm-test"))

	a, _ := queue.Enqueue[*User](q)
	b, _ := queue.Enqueue[*Note](q)
	c, _ := queue.Enqueue[*Like](q)

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	a <- &User{ID: "user", Text: "some text"}

	b <- &Note{ID: "note", Text: "some text"}

	c <- &Like{ID: "like", Text: "some text"}

	sys.Close()
}
