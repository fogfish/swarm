//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

func TestDequeuer(t *testing.T) {
	var bag []swarm.Bag
	bridge := &bridge{kernel.NewBridge(100 * time.Millisecond)}

	t.Run("New", func(t *testing.T) {
		q, err := NewDequeuer(
			WithConfig(
				swarm.WithLogStdErr(),
			),
		)
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Dequeue", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
			for _, m := range bag {
				bridge.Ack(context.Background(), m.Digest)
			}
		}()

		_, err := bridge.run(
			events.APIGatewayWebsocketProxyRequest{
				RequestContext: events.APIGatewayWebsocketProxyRequestContext{
					RouteKey:     "test",
					ConnectionID: "digest",
				},
				Body: `{"sut":"test"}`,
			},
		)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(len(bag), 1),
			it.Equal(bag[0].Category, "test"),
			it.Equal(bag[0].Digest, "digest"),
			it.Equiv(bag[0].Object, []byte(`{"sut":"test"}`)),
		)
	})
}
