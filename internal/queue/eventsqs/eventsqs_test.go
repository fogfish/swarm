//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/fogfish/swarm"
	sutRecv "github.com/fogfish/swarm/internal/queue/eventsqs"
	"github.com/fogfish/swarm/internal/queue/qtest"
	sutSend "github.com/fogfish/swarm/internal/queue/sqs"
	"github.com/fogfish/swarm/internal/system"
)

func mkQueue(sys swarm.System, policy *swarm.Policy, eff chan string) (swarm.Enqueue, swarm.Dequeue) {
	awscli, err := system.NewSession()
	if err != nil {
		panic(err)
	}

	enq := sutSend.NewEnqueue(sys, "test-sqs", policy, awscli)
	enq.Mock(&mockSQS{loopback: eff})

	deq := sutRecv.NewDequeue(sys, "test-bridge", policy)
	deq.Mock(mockLambda(eff))
	return enq, deq
}

func TestEventSQS(t *testing.T) {
	qtest.TestEnqueue(t, mkQueue)
	qtest.TestDequeue(t, mkQueue)
}

//
//
type mockSQS struct {
	sqsiface.SQSAPI

	loopback chan string
}

func (m *mockSQS) GetQueueUrl(s *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/000000000000/mock"),
	}, nil
}

func (m *mockSQS) SendMessage(s *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	cat, exists := s.MessageAttributes["Category"]
	if !exists {
		return nil, fmt.Errorf("Bad message attributes")
	}

	if !strings.HasPrefix(*cat.StringValue, qtest.Category) {
		return nil, fmt.Errorf("Bad message category")
	}

	m.loopback <- aws.StringValue(s.MessageBody)
	return &sqs.SendMessageOutput{}, nil
}

/*

mock AWS Lambda Handler

*/
func mockLambda(loopback chan string) func(interface{}) {
	return func(handler interface{}) {
		msg, _ := json.Marshal(
			events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId:     "abc-def",
						ReceiptHandle: qtest.Receipt,
						Body:          qtest.Message,
						MessageAttributes: map[string]events.SQSMessageAttribute{
							"Category": {StringValue: aws.String(qtest.Category)},
						},
					},
				},
			},
		)

		h := lambda.NewHandler(handler)
		_, err := h.Invoke(context.Background(), msg)
		if err != nil {
			panic(err)
		}

		loopback <- qtest.Receipt
	}
}