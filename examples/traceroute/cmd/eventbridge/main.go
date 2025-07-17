//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/examples/traceroute/core"
	"github.com/fogfish/swarm/examples/traceroute/pipe"
	"github.com/fogfish/swarm/listen"
)

func main() {
	k := eventbridge.Must(
		eventbridge.Listener().Build(),
	)

	go recv(listen.Event[core.ReqTrace](k))

	k.Await()
}

func recv(rcv <-chan core.ReqTrace, ack chan<- core.ReqTrace) {
	status := pipe.ActStatus()

	for evt := range rcv {
		req := core.ToRequest(evt)

		status <- core.GetStatus(req)
		if err := pipe.ToDynamoDB(evt); err != nil {
			ack <- evt.Fail(err)
			continue
		}

		ack <- evt
	}
}
