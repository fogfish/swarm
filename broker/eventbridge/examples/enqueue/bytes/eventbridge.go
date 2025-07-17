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
)

func main() {
	q := eventbridge.Must(
		eventbridge.Emitter().
			WithKernel(swarm.WithAgent("eventbridge:example/bytes")).
			Build("swarm-example-eventbridge"),
	)

	user := swarm.LogDeadLetters(eventbridge.EmitBytes(q, "User"))
	note := swarm.LogDeadLetters(eventbridge.EmitBytes(q, "Note"))
	like := swarm.LogDeadLetters(eventbridge.EmitBytes(q, "Like"))

	user <- []byte(`User Signed in`)
	note <- []byte(`User wrote note`)
	like <- []byte(`User liked note`)

	q.Close()
}
