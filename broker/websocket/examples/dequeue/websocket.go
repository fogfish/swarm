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

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/websocket"
	"github.com/fogfish/swarm/queue"
)

type User struct {
	Action string `json:"action"`
	ID     string `json:"id"`
	Text   string `json:"text"`
}

func main() {
	q := queue.Must(websocket.New(os.Getenv("CONFIG_SWARM_WS_URL"), swarm.WithLogStdErr()))

	a := &actor{emit: queue.New[User](q)}

	go a.handle(queue.Dequeue[User](q))

	q.Await()
}

type actor struct {
	emit queue.Enqueuer[User]
}

func (a *actor) handle(rcv <-chan swarm.Msg[User], ack chan<- swarm.Msg[User]) {
	for msg := range rcv {
		ctx := msg.Ctx.Value(websocket.WSRequest).(events.APIGatewayWebsocketProxyRequestContext)

		slog.Info("Event user", "msg", msg.Object, "ctx", ctx)

		if err := a.emit.Enq(msg.Ctx.Digest, msg.Object); err != nil {
			ack <- msg.Fail(err)
			continue
		}

		ack <- msg
	}
}
