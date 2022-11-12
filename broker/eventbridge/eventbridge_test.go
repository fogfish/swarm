//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge_test

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
	"github.com/fogfish/swarm/internal/qtest"
	"github.com/fogfish/swarm/queue"
)

func TestEnqueueSQS(t *testing.T) {
	qtest.TestEnqueueTyped(t, newMockEnqueue)
	qtest.TestEnqueueBytes(t, newMockEnqueue)
	qtest.TestEnqueueEvent(t, newMockEnqueue)
}

func TestDequeueSQS(t *testing.T) {
	qtest.TestDequeueTyped(t, newMockDequeue)
	qtest.TestDequeueBytes(t, newMockDequeue)
	qtest.TestDequeueEvent(t, newMockDequeue)
}

//
// Mock AWS EventBridge Enqueue
//
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

//
// Mock AWS EventBridge Dequeue
//
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
