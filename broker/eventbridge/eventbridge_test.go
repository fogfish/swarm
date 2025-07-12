//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

func TestReader(t *testing.T) {
	var bag []swarm.Bag
	cfg := swarm.NewConfig()
	cfg.TimeToFlight = 100 * time.Millisecond
	bridge := &bridge{kernel.NewBridge(cfg)}

	t.Run("New", func(t *testing.T) {
		_, err := Listener().Build()
		it.Then(t).Should(it.Nil(err))
	})

	t.Run("Dequeue", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
			for _, m := range bag {
				bridge.Ack(context.Background(), m.Digest)
			}
		}()

		err := bridge.run(context.Background(),
			events.CloudWatchEvent{
				ID:         "abc-def",
				DetailType: "category",
				Detail:     json.RawMessage(`{"sut":"test"}`),
			},
		)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(len(bag), 1),
			it.Equal(bag[0].Category, "category"),
			it.Equiv(bag[0].Object, []byte(`{"sut":"test"}`)),
		)
	})

	t.Run("Dequeue.Timeout", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
		}()

		err := bridge.run(context.Background(),
			events.CloudWatchEvent{
				ID:         "abc-def",
				DetailType: "category",
				Detail:     json.RawMessage(`{"sut":"test"}`),
			},
		)

		it.Then(t).ShouldNot(
			it.Nil(err),
		)
	})

	t.Run("Dequeue.EmptyDetail", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
			for _, m := range bag {
				bridge.Ack(context.Background(), m.Digest)
			}
		}()

		err := bridge.run(context.Background(),
			events.CloudWatchEvent{
				ID:         "abc-def",
				DetailType: "category",
				Detail:     json.RawMessage(`{}`),
			},
		)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(len(bag), 1),
			it.Equal(bag[0].Category, "category"),
			it.Equiv(bag[0].Object, []byte(`{}`)),
		)
	})

	t.Run("Dequeue.ErrorAck", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
			for _, m := range bag {
				bridge.Err(context.Background(), m.Digest, fmt.Errorf("processing failed"))
			}
		}()

		err := bridge.run(context.Background(),
			events.CloudWatchEvent{
				ID:         "test-error",
				DetailType: "category",
				Detail:     json.RawMessage(`{"error":"test"}`),
			},
		)

		it.Then(t).ShouldNot(
			it.Nil(err),
		)
	})
}

func TestWriter(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		q, err := Emitter().Build("test")
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Enqueue", func(t *testing.T) {
		mock := &mockEventBridge{}

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
			it.Equal(*mock.val.DetailType, "cat"),
			it.Equal(*mock.val.Detail, "value"),
		)

		q.Close()
	})

	t.Run("Enqueue.ServiceError", func(t *testing.T) {
		mock := &mockEventBridgeError{}

		q, err := Emitter().WithService(mock).Build("test")
		it.Then(t).Should(it.Nil(err))

		err = q.Emitter.Enq(context.Background(),
			swarm.Bag{
				Category: "cat",
				Object:   []byte(`value`),
			},
		)
		it.Then(t).ShouldNot(it.Nil(err))

		q.Close()
	})

	t.Run("Enqueue.FailedEntry", func(t *testing.T) {
		mock := &mockEventBridgeFailedEntry{}

		q, err := Emitter().WithService(mock).Build("test")
		it.Then(t).Should(it.Nil(err))

		err = q.Emitter.Enq(context.Background(),
			swarm.Bag{
				Category: "cat",
				Object:   []byte(`value`),
			},
		)
		it.Then(t).ShouldNot(it.Nil(err))

		q.Close()
	})

	t.Run("Enqueue.ContextCancel", func(t *testing.T) {
		mock := &mockEventBridgeTimeout{}

		q, err := Emitter().WithService(mock).Build("test")
		it.Then(t).Should(it.Nil(err))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err = q.Emitter.Enq(ctx,
			swarm.Bag{
				Category: "cat",
				Object:   []byte(`value`),
			},
		)
		it.Then(t).ShouldNot(it.Nil(err))

		q.Close()
	})
}

func TestBroker(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		q, err := Endpoint().Build("test")
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

}

func TestConfig(t *testing.T) {
	t.Run("WithEventBus", func(t *testing.T) {
		q, err := Emitter().Build("test-bus")
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("WithService", func(t *testing.T) {
		mock := &mockEventBridge{}
		q, err := Emitter().WithService(mock).Build("test-bus")
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("MustEnqueuer.Success", func(t *testing.T) {
		q := Must(Emitter().Build("test"))
		it.Then(t).ShouldNot(it.Nil(q))
		q.Close()
	})

	t.Run("MustDequeuer.Success", func(t *testing.T) {
		q := Must(Listener().Build())
		it.Then(t).ShouldNot(it.Nil(q))
		q.Close()
	})
}

func TestBuilder(t *testing.T) {
	t.Run("Simple case with event bus", func(t *testing.T) {
		mock := &mockEventBridge{}
		client, err := Endpoint().
			WithService(mock).
			Build("production-events")

		it.Then(t).Should(it.Nil(err))
		client.Close()
	})

	t.Run("Enqueuer with kernel options", func(t *testing.T) {
		mock := &mockEventBridge{}
		enqueuer, err := Emitter().
			WithKernel(swarm.WithAgent("test-service")).
			WithService(mock).
			Build("production-events")

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(enqueuer.Config.Agent, "test-service"),
		)
		enqueuer.Close()
	})
}

//------------------------------------------------------------------------------

type mockEventBridge struct {
	EventBridge
	val types.PutEventsRequestEntry
}

func (m *mockEventBridge) PutEvents(ctx context.Context, req *eventbridge.PutEventsInput, opts ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	if len(req.Entries) != 1 {
		return nil, fmt.Errorf("Bad request")
	}

	m.val = req.Entries[0]

	return &eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
	}, nil
}

type mockEventBridgeError struct {
	EventBridge
}

func (m *mockEventBridgeError) PutEvents(ctx context.Context, req *eventbridge.PutEventsInput, opts ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	return nil, fmt.Errorf("network error")
}

type mockEventBridgeFailedEntry struct {
	EventBridge
}

func (m *mockEventBridgeFailedEntry) PutEvents(ctx context.Context, req *eventbridge.PutEventsInput, opts ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	return &eventbridge.PutEventsOutput{
		FailedEntryCount: 1,
		Entries: []types.PutEventsResultEntry{
			{
				ErrorCode:    aws.String("ValidationException"),
				ErrorMessage: aws.String("Event size exceeds maximum"),
			},
		},
	}, nil
}

type mockEventBridgeTimeout struct {
	EventBridge
}

func (m *mockEventBridgeTimeout) PutEvents(ctx context.Context, req *eventbridge.PutEventsInput, opts ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(100 * time.Millisecond):
		return &eventbridge.PutEventsOutput{
			FailedEntryCount: 0,
		}, nil
	}
}
