//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"log/slog"
	"os"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/websocket"
	"github.com/fogfish/swarm/listen"
)

type UserEvent = swarm.Event[swarm.Meta, User]

type User struct {
	Action string `json:"action"`
	ID     string `json:"id"`
	Text   string `json:"text"`
}

func main() {
	q := websocket.Must(
		websocket.Endpoint().Build(
			os.Getenv(websocket.EnvConfigSourceWebSocketUrl),
		),
	)

	emit := swarm.LogDeadLetters(websocket.EmitEvent[UserEvent](q.Emitter))

	a := &actor{emit: emit}
	go a.handle(listen.Event[UserEvent](q.Listener))

	q.Await()
}

type actor struct {
	emit chan<- UserEvent
}

func (a *actor) handle(rcv <-chan UserEvent, ack chan<- UserEvent) {
	for msg := range rcv {
		slog.Info("Event user", "evt", msg)

		a.emit <- msg

		ack <- msg
	}
}
