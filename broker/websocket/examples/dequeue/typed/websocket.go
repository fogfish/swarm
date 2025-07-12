//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/websocket"
	"github.com/fogfish/swarm/dequeue"
	"github.com/fogfish/swarm/enqueue"
)

type User struct {
	Action string `json:"action"`
	ID     string `json:"id"`
	Text   string `json:"text"`
}

func main() {
	q := websocket.Channels().MustClient(
		os.Getenv(websocket.EnvConfigSourceWebSocket),
	)

	a := &actor{emit: enqueue.NewTyped[User](q.EmitterCore)}
	go a.handle(dequeue.Typed[User](q.ListenerCore))

	q.Await()
}

type actor struct {
	emit *enqueue.EmitterTyped[User]
}

func (a *actor) handle(rcv <-chan swarm.Msg[User], ack chan<- swarm.Msg[User]) {
	for msg := range rcv {
		slog.Info("Event user", "data", msg.Object)

		if err := a.emit.Enq(context.Background(), msg.Object, msg.Digest); err != nil {
			ack <- msg.Fail(err)
			continue
		}

		ack <- msg
	}
}
