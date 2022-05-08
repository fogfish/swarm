//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"fmt"

	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/sqs"
)

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func main() {
	sys := sqs.NewSystem("swarm-example-sqs")
	q := sqs.Must(sqs.New(sys, "swarm-test"))

	a, _ := queue.Send[*Note](q)

	// a, _ := q.Send("sqs.test.a")
	// b, _ := q.Send("sqs.test.b")
	// c, _ := q.Send("sqs.test.c")

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	fmt.Println(a)

	a <- &Note{ID: "a", Text: "message"}
	// fmt.Println(<-e)

	// time.Sleep(10 * time.Second)
	// a <- swarm.Bytes("{\"type\": \"a\", \"some\": \"message\"}")
	// b <- swarm.Bytes("{\"type\": \"b\", \"some\": \"message\"}")
	// c <- swarm.Bytes("{\"type\": \"c\", \"some\": \"message\"}")

	sys.Stop()
}
