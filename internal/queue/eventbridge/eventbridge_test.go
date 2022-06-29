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
	sut "github.com/fogfish/swarm/internal/queue/eventbridge"
	"github.com/fogfish/swarm/internal/queue/qtest"
)

func TestEventBridge(t *testing.T) {
	qtest.TestEnqueue(t, mkEnqueue)
	qtest.TestDequeue(t, mkDequeue)
}

func mkEnqueue(
	sys swarm.System,
	policy *swarm.Policy,
	eff chan string,
	expectCategory string,
) swarm.Enqueue {
	enq := sut.NewEnqueue(sys, "test-bridge", policy, *aws.NewConfig())
	enq.Mock(&mockEnqueueService{
		loopback:       eff,
		expectCategory: expectCategory,
	})

	return enq
}

func mkDequeue(
	sys swarm.System,
	policy *swarm.Policy,
	eff chan string,
	returnCategory string,
	returnMessage string,
	returnReceipt string,
) swarm.Dequeue {
	deq := sut.NewDequeue(sys, "test-bridge", policy)
	deq.Mock(mockLambda(eff, returnCategory, returnMessage, returnReceipt))
	return deq
}

type mockEnqueueService struct {
	sut.EnqueueService
	loopback       chan string
	expectCategory string
}

func (m *mockEnqueueService) PutEvents(ctx context.Context, req *eventbridge.PutEventsInput, opts ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
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

/*

mock AWS Lambda Handler

*/
func mockLambda(
	loopback chan string,
	returnCategory string,
	returnMessage string,
	returnReceipt string,
) func(interface{}) {
	return func(handler interface{}) {
		msg, _ := json.Marshal(
			events.CloudWatchEvent{
				ID:         "abc-def",
				DetailType: returnCategory,
				Detail:     json.RawMessage(returnMessage),
			},
		)

		h := lambda.NewHandler(handler)
		_, err := h.Invoke(context.Background(), msg)
		if err != nil {
			panic(err)
		}

		loopback <- returnReceipt
	}
}
