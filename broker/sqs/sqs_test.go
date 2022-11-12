//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package sqs_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fogfish/swarm"
	sut "github.com/fogfish/swarm/broker/sqs"
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
// Mock AWS SQS Enqueue
//
type mockEnqueue struct {
	sut.SQS
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

func (m *mockEnqueue) GetQueueUrl(ctx context.Context, req *sqs.GetQueueUrlInput, opts ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
	return &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/000000000000/mock"),
	}, nil
}

func (m *mockEnqueue) SendMessage(ctx context.Context, req *sqs.SendMessageInput, opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	cat, exists := req.MessageAttributes["Category"]
	if !exists {
		return nil, fmt.Errorf("Bad message attributes")
	}

	if !strings.HasPrefix(*cat.StringValue, m.expectCategory) {
		return nil, fmt.Errorf("Bad message category")
	}

	m.loopback <- aws.ToString(req.MessageBody)
	return &sqs.SendMessageOutput{}, nil
}

//
// Mock AWS SQS Dequeue
//
type mockDequeue struct {
	sut.SQS
	returnCategory string
	returnMessage  string
	returnReceipt  string
	loopback       chan string
}

func newMockDequeue(
	loopback chan string,
	queueName string,
	returnCategory string,
	returnMessage string,
	returnReceipt string,
	opts ...swarm.Option,
) swarm.Broker {
	mock := &mockDequeue{
		loopback:       loopback,
		returnCategory: returnCategory,
		returnMessage:  returnMessage,
		returnReceipt:  returnReceipt,
	}
	conf := append(opts, swarm.WithService(mock))
	return queue.Must(sut.New(queueName, conf...))
}

func (m *mockDequeue) GetQueueUrl(ctx context.Context, req *sqs.GetQueueUrlInput, opts ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
	return &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/000000000000/mock"),
	}, nil
}

func (m *mockDequeue) ReceiveMessage(ctx context.Context, req *sqs.ReceiveMessageInput, opts ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	return &sqs.ReceiveMessageOutput{
		Messages: []types.Message{
			{
				MessageAttributes: map[string]types.MessageAttributeValue{
					"Category": {StringValue: aws.String(m.returnCategory)},
				},
				Body:          aws.String(m.returnMessage),
				ReceiptHandle: aws.String(m.returnReceipt),
			},
		},
	}, nil
}

func (m *mockDequeue) DeleteMessage(ctx context.Context, req *sqs.DeleteMessageInput, opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	m.loopback <- aws.ToString(req.ReceiptHandle)
	return &sqs.DeleteMessageOutput{}, nil
}
