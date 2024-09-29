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
		q, err := NewReader("test")
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
}

func TestWriter(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		q, err := NewWriter("test")
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Enqueue", func(t *testing.T) {
		mock := &mockEventBridge{}

		q, err := NewWriter("test", WithService(mock))
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
