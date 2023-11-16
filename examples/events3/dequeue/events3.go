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

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/events3"
	"github.com/fogfish/swarm/queue"
)

func main() {
	q := queue.Must(events3.New("swarm-test", swarm.WithLogStdErr()))

	go common(events3.Dequeue(q))

	q.Await()
}

func common(rcv <-chan *events3.Event, ack chan<- *events3.Event) {
	for msg := range rcv {

		v, _ := json.MarshalIndent(msg, "", " ")
		fmt.Printf("s3 event > \n %s\n", v)
		ack <- msg
	}
}
