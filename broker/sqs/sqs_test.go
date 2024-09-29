//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package sqs_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/dequeue"
)

func TestEnqueuer(t *testing.T) {
	t.Run("NewEnqueuer", func(t *testing.T) {
		mock := &mockEnqueue{}
		q, err := sqs.NewEnqueuer("test", sqs.WithService(mock))
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Enqueue", func(t *testing.T) {
		mock := &mockEnqueue{}

		q, err := sqs.NewEnqueuer("test", sqs.WithService(mock))
		it.Then(t).Should(it.Nil(err))

		err = q.Emitter.Enq(context.Background(),
			swarm.Bag{
				Category: "cat",
				Object:   []byte(`value`),
			},
		)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(*mock.req.MessageAttributes["Category"].StringValue, "cat"),
			it.Equal(*mock.req.MessageBody, "value"),
		)

		q.Close()
	})
}

func TestDequeuer(t *testing.T) {
	t.Run("NewDequeuer", func(t *testing.T) {
		mock := &mockDequeue{}
		q, err := sqs.NewDequeuer("test", sqs.WithService(mock))
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Dequeue", func(t *testing.T) {
		mock := &mockDequeue{}

		q, err := sqs.NewDequeuer("test", sqs.WithService(mock))
		it.Then(t).Should(it.Nil(err))

		rcv, ack := dequeue.Bytes(q, "test")
		go func() {
			ack <- <-rcv

			time.Sleep(5 * time.Millisecond)
			q.Close()
		}()

		q.Await()

		it.Then(t).Should(
			it.Equal(*mock.req.ReceiptHandle, "1"),
		)
	})
}

//------------------------------------------------------------------------------

type mockEnqueue struct {
	sqs.SQS
	req *awssqs.SendMessageInput
}

func (m *mockEnqueue) GetQueueUrl(ctx context.Context, req *awssqs.GetQueueUrlInput, opts ...func(*awssqs.Options)) (*awssqs.GetQueueUrlOutput, error) {
	return &awssqs.GetQueueUrlOutput{
		QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/000000000000/mock"),
	}, nil
}

func (m *mockEnqueue) SendMessage(ctx context.Context, req *awssqs.SendMessageInput, opts ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error) {
	m.req = req
	return &awssqs.SendMessageOutput{}, nil
}

type mockDequeue struct {
	sqs.SQS
	req *awssqs.DeleteMessageInput
}

func (m *mockDequeue) GetQueueUrl(ctx context.Context, req *awssqs.GetQueueUrlInput, opts ...func(*awssqs.Options)) (*awssqs.GetQueueUrlOutput, error) {
	return &awssqs.GetQueueUrlOutput{
		QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/000000000000/mock"),
	}, nil
}

func (m *mockDequeue) ReceiveMessage(ctx context.Context, req *awssqs.ReceiveMessageInput, opts ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error) {
	return &awssqs.ReceiveMessageOutput{
		Messages: []types.Message{
			{
				MessageAttributes: map[string]types.MessageAttributeValue{
					"Category": {StringValue: aws.String("test")},
				},
				Body:          aws.String(`{"sut":"test"}`),
				ReceiptHandle: aws.String(`1`),
			},
		},
	}, nil
}

func (m *mockDequeue) DeleteMessage(ctx context.Context, req *awssqs.DeleteMessageInput, opts ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error) {
	m.req = req
	return &awssqs.DeleteMessageOutput{}, nil
}
