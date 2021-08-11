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
	"github.com/fogfish/swarm/queue/sqs"
)

func main() {
	sys := swarm.New("test")
	queue := swarm.Must(sqs.New(sys, "swarm-test"))

	a, _ := queue.Send("sqs.test.a")
	b, _ := queue.Send("sqs.test.b")
	c, _ := queue.Send("sqs.test.c")

	a <- swarm.Bytes("{\"type\": \"a\", \"some\": \"message\"}")
	b <- swarm.Bytes("{\"type\": \"b\", \"some\": \"message\"}")
	c <- swarm.Bytes("{\"type\": \"c\", \"some\": \"message\"}")

	sys.Stop()
}
