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
	bridge := &bridge{kernel.NewBridge(100 * time.Millisecond)}

	t.Run("New", func(t *testing.T) {
		_, err := NewDequeuer()
		it.Then(t).Should(it.Nil(err))
	})

	t.Run("Dequeue", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
			for _, m := range bag {
				bridge.Ack(context.Background(), m.Digest)
			}
		}()

		err := bridge.run(
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

		err := bridge.run(
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

		err := bridge.run(
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

		err := bridge.run(
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
		q, err := NewEnqueuer(WithEventBus("test"))
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("New.NoEventBus", func(t *testing.T) {
		_, err := NewEnqueuer()
		it.Then(t).ShouldNot(it.Nil(err))
	})

	t.Run("Enqueue", func(t *testing.T) {
		mock := &mockEventBridge{}

		q, err := NewEnqueuer(WithEventBus("test"), WithService(mock))
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

		q, err := NewEnqueuer(WithEventBus("test"), WithService(mock))
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

		q, err := NewEnqueuer(WithEventBus("test"), WithService(mock))
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

		q, err := NewEnqueuer(WithEventBus("test"), WithService(mock))
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
		q, err := New(WithEventBus("test"))
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("New.NoEventBus", func(t *testing.T) {
		// EventBus is now mandatory - should return an error
		_, err := New()
		it.Then(t).ShouldNot(it.Nil(err))
	})
}

func TestConfig(t *testing.T) {
	t.Run("WithEventBus", func(t *testing.T) {
		q, err := NewEnqueuer(WithEventBus("test-bus"))
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("WithService", func(t *testing.T) {
		mock := &mockEventBridge{}
		q, err := NewEnqueuer(WithEventBus("test"), WithService(mock))
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("MustEnqueuer.Success", func(t *testing.T) {
		q := MustEnqueuer(WithEventBus("test"))
		it.Then(t).ShouldNot(it.Nil(q))
		q.Close()
	})

	t.Run("MustDequeuer.Success", func(t *testing.T) {
		q := MustDequeuer(WithEventBus("test-bus"))
		it.Then(t).ShouldNot(it.Nil(q))
		q.Close()
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
