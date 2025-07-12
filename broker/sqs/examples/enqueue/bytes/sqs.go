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
	enqueue "github.com/fogfish/swarm/emit"
	"github.com/fogfish/swarm/kernel/encoding"
)

func main() {
	q := sqs.Must(sqs.Emitter().Build("swarm-test"))

	user := swarm.LogDeadLetters(enqueue.Bytes(q, encoding.ForBytes("User")))
	note := swarm.LogDeadLetters(enqueue.Bytes(q, encoding.ForBytes("Note")))
	like := swarm.LogDeadLetters(enqueue.Bytes(q, encoding.ForBytes("Like")))

	user <- []byte("user|some text by user")

	note <- []byte("note|some note")

	like <- []byte("like|someone liked it")

	q.Close()
}
