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
	"github.com/fogfish/swarm/internal/qtest"
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/bytes"
)

func main() {
	qtest.NewLogger()

	q := queue.Must(sqs.New("swarm-test", swarm.WithLogStdErr()))

	user := swarm.LogDeadLetters(bytes.Enqueue(q, "User"))
	note := swarm.LogDeadLetters(bytes.Enqueue(q, "Note"))
	like := swarm.LogDeadLetters(bytes.Enqueue(q, "Like"))

	user <- []byte("user|some text by user")

	note <- []byte("note|some note")

	like <- []byte("like|someone liked it")

	q.Close()
}
