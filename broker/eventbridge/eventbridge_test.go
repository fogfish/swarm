//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
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

/*

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/fogfish/swarm"
	sut "github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/qtest"
	"github.com/fogfish/swarm/queue"
)


func TestEnqueueEventBridge(t *testing.T) {
	qtest.TestEnqueueTyped(t, newMockEnqueue)
	qtest.TestEnqueueBytes(t, newMockEnqueue)
	qtest.TestEnqueueEvent(t, newMockEnqueue)
}

func TestDequeueEventBridge(t *testing.T) {
	qtest.TestDequeueTyped(t, newMockDequeue)
	qtest.TestDequeueBytes(t, newMockDequeue)
	qtest.TestDequeueEvent(t, newMockDequeue)
}

// Mock AWS EventBridge Enqueue
type mockEnqueue struct {
	sut.EventBridge
	expectCategory string
	loopback       chan string
}

func newMockEnqueue(
	loopback chan string,
	queueName string,
	expectCategory string,
	opts ...swarm.Option,
) swarm.Broker {
	mock := &mockEnqueue{expectCategory: expectCategory, loopback: loopback}
	conf := append(opts, swarm.WithService(mock))
	return queue.Must(sut.New(queueName, conf...))
}

func (m *mockEnqueue) PutEvents(ctx context.Context, req *eventbridge.PutEventsInput, opts ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	if len(req.Entries) != 1 {
		return nil, fmt.Errorf("Bad request")
	}

	if !strings.HasPrefix(*req.Entries[0].DetailType, m.expectCategory) {
		return nil, fmt.Errorf("Bad message category")
	}

	m.loopback <- aws.ToString(req.Entries[0].Detail)
	return &eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
	}, nil
}

// Mock AWS EventBridge Dequeue
func newMockDequeue(
	loopback chan string,
	queueName string,
	returnCategory string,
	returnMessage string,
	returnReceipt string,
	opts ...swarm.Option,
) swarm.Broker {
	mock := &mockLambda{
		loopback:       loopback,
		returnCategory: returnCategory,
		returnMessage:  returnMessage,
		returnReceipt:  returnReceipt,
	}
	conf := append(opts, swarm.WithService(mock))
	return queue.Must(sut.New(queueName, conf...))
}

type mockLambda struct {
	sut.EventBridge
	loopback       chan string
	returnCategory string
	returnMessage  string
	returnReceipt  string
}

func (mock *mockLambda) Start(handler interface{}) {
	msg, _ := json.Marshal(
		events.CloudWatchEvent{
			ID:         "abc-def",
			DetailType: mock.returnCategory,
			Detail:     json.RawMessage(mock.returnMessage),
		},
	)

	h := lambda.NewHandler(handler)
	_, err := h.Invoke(context.Background(), msg)
	if err != nil {
		panic(err)
	}

	mock.loopback <- mock.returnReceipt
}

*/

func TestReader(t *testing.T) {
	var bag []swarm.Bag
	bridge := &bridge{kernel.NewBridge(100 * time.Millisecond)}

	t.Run("Success", func(t *testing.T) {
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

	t.Run("Timeout", func(t *testing.T) {
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
