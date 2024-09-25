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
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/fogfish/swarm"
	sut "github.com/fogfish/swarm/broker/eventsqs"
	sutsqs "github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/qtest"
	"github.com/fogfish/swarm/queue"
)

func TestDequeueEventSQS(t *testing.T) {
	qtest.TestDequeueTyped(t, newMockDequeue)
	qtest.TestDequeueBytes(t, newMockDequeue)
	qtest.TestDequeueEvent(t, newMockDequeue)
}

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
	sutsqs.SQS
	loopback       chan string
	returnCategory string
	returnMessage  string
	returnReceipt  string
}

func (mock *mockLambda) GetQueueUrl(ctx context.Context, req *sqs.GetQueueUrlInput, opts ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
	return &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/000000000000/mock"),
	}, nil
}

func (mock *mockLambda) Start(handler interface{}) {
	msg, _ := json.Marshal(
		events.SQSEvent{
			Records: []events.SQSMessage{
				{
					MessageId:     "abc-def",
					ReceiptHandle: mock.returnReceipt,
					Body:          mock.returnMessage,
					MessageAttributes: map[string]events.SQSMessageAttribute{
						"Category": {StringValue: aws.String(mock.returnCategory)},
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

	mock.loopback <- mock.returnReceipt
}
