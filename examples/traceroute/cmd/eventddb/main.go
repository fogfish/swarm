//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventddb"
	"github.com/fogfish/swarm/examples/traceroute/core"
	"github.com/fogfish/swarm/examples/traceroute/pipe"
)

func main() {
	k := eventddb.Must(
		eventddb.Listener().Build(),
	)

	go recv(eventddb.Listen(k))

	k.Await()
}

func recv(rcv <-chan swarm.Msg[*events.DynamoDBEventRecord], ack chan<- swarm.Msg[*events.DynamoDBEventRecord]) {
	status := pipe.ActStatus()

	for evt := range rcv {
		if evt.Object.EventName != "INSERT" {
			ack <- evt
			continue
		}

		var rdb core.RequestDB
		if err := eventddb.UnmarshalMap(evt.Object.Change.NewImage, &rdb); err != nil {
			slog.Error("failed to unmarshal request", "err", err)
			ack <- evt.Fail(err)
			continue
		}

		req := core.FromRequestDB(rdb)
		status <- core.GetStatus(core.ToRequest(req))
		if err := pipe.ToS3(req); err != nil {
			ack <- evt.Fail(err)
			continue
		}

		ack <- evt
	}
}
