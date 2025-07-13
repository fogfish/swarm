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

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/embedded"
	"github.com/fogfish/swarm/emit"
	"github.com/fogfish/swarm/listen"
)

// Example message
type Order struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

func main() {
	// create broker
	q := embedded.Must(embedded.Endpoint().Build())

	// create Golang channels
	rcv, ack := listen.Typed[Order](q.ListenerCore)
	out := swarm.LogDeadLetters(emit.Typed[Order](q.EmitterCore))

	// Send messages
	out <- Order{ID: "123", Amount: 100.0}

	// use Golang channels for I/O
	go func() {
		for order := range rcv {
			processOrder(order.Object)
			ack <- order // acknowledge processing
		}
	}()

	q.Await()
}

func processOrder(order Order) {
	slog.Info("processing order", "order", order)
	// Your business logic here
}
