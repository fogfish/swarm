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

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventddb"
)

func main() {
	q := eventddb.Channels().MustDequeuer()

	go common(eventddb.Source(q))

	q.Await()
}

func common(rcv <-chan swarm.Msg[*events.DynamoDBEventRecord], ack chan<- swarm.Msg[*events.DynamoDBEventRecord]) {
	for msg := range rcv {
		v, _ := json.MarshalIndent(msg, "", " ")
		fmt.Printf("ddb event > \n %s\n", v)
		ack <- msg
	}
}
