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
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

func TestDequeuer(t *testing.T) {
	var bag []swarm.Bag
	cfg := swarm.NewConfig()
	cfg.TimeToFlight = 100 * time.Millisecond
	bridge := &bridge{kernel.NewBridge(cfg)}

	// Test new builder pattern
	t.Run("Builder.NewDequeuer", func(t *testing.T) {
		q, err := Listener().Build()
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Builder.WithKernel", func(t *testing.T) {
		q, err := Listener().
			WithKernel(swarm.WithLogStdErr()).
			Build()
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

		_, err := bridge.run(context.Background(),
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
		).ShouldNot(
			it.Nil(bag[0].IOContext),
		)

		ctx := bag[0].IOContext.(*events.APIGatewayWebsocketProxyRequestContext)
		it.Then(t).Should(
			it.Equal(ctx.RouteKey, "test"),
			it.Equal(ctx.ConnectionID, "digest"),
		)
	})
}

func TestEnqueuer(t *testing.T) {
	// Test new builder pattern
	t.Run("Builder.NewEnqueuer", func(t *testing.T) {
		q, err := Emitter().Build("test")
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Builder.WithService", func(t *testing.T) {
		mock := &mockGateway{}
		q, err := Emitter().
			WithService(mock).
			Build("test")
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Builder.NewClient", func(t *testing.T) {
		mock := &mockGateway{}
		q, err := Endpoint().
			WithService(mock).
			Build("test")
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Enqueue", func(t *testing.T) {
		mock := &mockGateway{}

		q, err := Emitter().WithService(mock).Build("test")
		it.Then(t).Should(it.Nil(err))

		err = q.Emitter.Enq(context.Background(),
			swarm.Bag{
				Category: "cat",
				Object:   []byte(`value`),
			},
		)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(*mock.req.ConnectionId, "cat"),
			it.Equal(string(mock.req.Data), "value"),
		)

		q.Close()
	})
}

//------------------------------------------------------------------------------

type mockGateway struct {
	Gateway
	req *apigatewaymanagementapi.PostToConnectionInput
}

func (m *mockGateway) PostToConnection(ctx context.Context, req *apigatewaymanagementapi.PostToConnectionInput, optFns ...func(*apigatewaymanagementapi.Options)) (*apigatewaymanagementapi.PostToConnectionOutput, error) {
	m.req = req

	return &apigatewaymanagementapi.PostToConnectionOutput{}, nil
}
