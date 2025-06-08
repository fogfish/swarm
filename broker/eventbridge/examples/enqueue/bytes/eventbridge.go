//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/enqueue"
	"github.com/fogfish/swarm/kernel/encoding"
)

func main() {
	q := eventbridge.MustEnqueuer(
		eventbridge.WithEventBus("swarm-example-eventbridge"),
		eventbridge.WithConfig(
			swarm.WithSource("swarm-example-eventbridge"),
			swarm.WithLogStdErr(),
		),
	)

	user := swarm.LogDeadLetters(enqueue.Bytes(q, encoding.ForBytesJB64("User")))
	note := swarm.LogDeadLetters(enqueue.Bytes(q, encoding.ForBytesJB64("Note")))
	like := swarm.LogDeadLetters(enqueue.Bytes(q, encoding.ForBytesJB64("Like")))

	user <- []byte(`User Signed in`)
	note <- []byte(`User wrote note`)
	like <- []byte(`User liked note`)

	q.Close()
}
