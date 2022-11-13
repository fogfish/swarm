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
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/bytes"
)

func main() {
	q := queue.Must(sqs.New("swarm-test"))

	user, _ := bytes.Enqueue(q, "User")
	note, _ := bytes.Enqueue(q, "Note")
	like, _ := bytes.Enqueue(q, "Like")

	user <- []byte("user|some text by user")

	note <- []byte("note|some note")

	like <- []byte("like|someone liked it")

	q.Close()
}
