//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/events3"
)

func main() {
	q, err := events3.NewDequeuer(
		events3.WithConfig(
			swarm.WithLogStdErr(),
		),
	)
	if err != nil {
		slog.Error("eventbridge reader has failed", "err", err)
		return
	}

	go common(events3.Source(q))

	q.Await()
}

func common(rcv <-chan swarm.Msg[*events.S3EventRecord], ack chan<- swarm.Msg[*events.S3EventRecord]) {
	for msg := range rcv {

		v, _ := json.MarshalIndent(msg, "", " ")
		fmt.Printf("s3 event > \n %s\n", v)
		ack <- msg
	}
}
