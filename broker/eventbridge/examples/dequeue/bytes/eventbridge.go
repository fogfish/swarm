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

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventbridge"
)

func main() {
	q := eventbridge.Must(eventbridge.Listener().Build())

	// Note: AWS EventBridge is not capable transmitting pure []byte, it works only with JSON.
	//       The broker implements eventbridge specific codec to encapsulate []byte int JSON before transmission.
	go actor("user").handle(eventbridge.RecvBytes(q, "User"))
	go actor("note").handle(eventbridge.RecvBytes(q, "Note"))
	go actor("like").handle(eventbridge.RecvBytes(q, "Like"))

	q.Await()
}

type actor string

func (a actor) handle(rcv <-chan swarm.Msg[[]byte], ack chan<- swarm.Msg[[]byte]) {
	for msg := range rcv {
		slog.Info("Event", "type", a, "msg", string(msg.Object))
		ack <- msg
	}
}
