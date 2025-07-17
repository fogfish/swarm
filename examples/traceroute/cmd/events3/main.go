//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/events3"
	"github.com/fogfish/swarm/examples/traceroute/core"
	"github.com/fogfish/swarm/examples/traceroute/pipe"
)

func main() {
	k := events3.Must(
		events3.Listener().Build(),
	)

	go recv(events3.Listen(k))

	k.Await()
}

func recv(rcv <-chan swarm.Msg[*events.S3EventRecord], ack chan<- swarm.Msg[*events.S3EventRecord]) {
	status := pipe.ActStatus()
	stream := pipe.ToSQS()

	for evt := range rcv {
		rdb, err := pipe.FromS3(evt.Object)
		if err != nil {
			ack <- evt.Fail(err)
			continue
		}

		req := core.FromRequestDB(rdb)
		status <- core.GetStatus(core.ToRequest(req))
		stream <- req

		ack <- evt
	}
}
