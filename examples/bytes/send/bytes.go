//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/fogfish/swarm/broker/sqs"
	queue "github.com/fogfish/swarm/queue/bytes"
)

func main() {
	q, err := sqs.New("swarm-test")
	if err != nil {
		panic(err)
	}

	user, _ := queue.Enqueue(q, "User")
	note, _ := queue.Enqueue(q, "Note")
	like, _ := queue.Enqueue(q, "Like")

	user <- []byte("user|some text by user")

	note <- []byte("note|some note")

	like <- []byte("like|someone liked it")

	q.Close()
}
