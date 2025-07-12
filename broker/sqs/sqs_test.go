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
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/kernel/encoding"
	dequeue "github.com/fogfish/swarm/listen"
)

func TestNew(t *testing.T) {
	mock := &mockEnqueue{}
	q, err := sqs.Emitter().
		WithService(mock).
		WithBatchSize(100).
		Build("test")

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(q.Config.PollerPool, 11),
	)
	q.Close()
}

func TestEnqueuer(t *testing.T) {
	t.Run("NewEnqueuer", func(t *testing.T) {
		mock := &mockEnqueue{}
		q, err := sqs.Emitter().
			WithService(mock).
			Build("test")

		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Enqueue", func(t *testing.T) {
		mock := &mockEnqueue{}
		q, err := sqs.Emitter().
			WithService(mock).
			Build("test")

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
		q, err := sqs.Listener().
			WithService(mock).
			WithBatchSize(10).
			Build("test")

		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("New", func(t *testing.T) {
		mock := &mockDequeue{}
		q, err := sqs.Listener().
			WithService(mock).
			WithBatchSize(10).
			Build("test")

		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Dequeue", func(t *testing.T) {
		mock := &mockDequeue{}

		q, err := sqs.Listener().
			WithService(mock).
			Build("test")

		it.Then(t).Should(it.Nil(err))

		rcv, ack := dequeue.Bytes(q, encoding.ForBytes("test"))
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

	t.Run("Dequeue.Error", func(t *testing.T) {
		mock := &mockDequeue{}

		q, err := sqs.Listener().WithService(mock).Build("test")
		it.Then(t).Should(it.Nil(err))

		rcv, ack := dequeue.Bytes(q, encoding.ForBytes("test"))
		go func() {
			msg := <-rcv
			ack <- msg.Fail(fmt.Errorf("fail"))

			time.Sleep(5 * time.Millisecond)
			q.Close()
		}()

		q.Await()

		it.Then(t).Should(
			it.True(mock.req == nil),
		)
	})
}

func TestBuilder(t *testing.T) {
	t.Run("Simple case with sensible defaults", func(t *testing.T) {
		mock := &mockEnqueue{}
		client, err := sqs.Endpoint().
			WithService(mock).
			Build("test")

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(client.EmitterCore.Config.Agent, "github.com/fogfish/swarm"), // Default source
			it.Equal(client.EmitterCore.Config.PollerPool, 1),                     // Default poller pool
		)
		client.Close()
	})

	t.Run("Broker customization", func(t *testing.T) {
		mock := &mockEnqueue{}
		enqueuer, err := sqs.Emitter().
			WithBatchSize(5).
			WithService(mock).
			Build("test")

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(enqueuer.Config.PollerPool, 1), // Batch size <= 10
		)
		enqueuer.Close()
	})

	t.Run("Large batch size adjusts poller pool", func(t *testing.T) {
		mock := &mockEnqueue{}
		dequeuer, err := sqs.Listener().
			WithBatchSize(25).
			WithService(mock).
			Build("test")

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(dequeuer.Config.PollerPool, 3), // 25/10 + 1 = 3
		)
		dequeuer.Close()
	})

	t.Run("Advanced kernel configuration", func(t *testing.T) {
		mock := &mockEnqueue{}
		client, err := sqs.Endpoint().
			WithKernel(
				swarm.WithAgent("advanced-service"),
				swarm.WithPollFrequency(500*time.Millisecond),
			).
			WithBatchSize(10).
			WithService(mock).
			Build("test")

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(client.EmitterCore.Config.Agent, "advanced-service"),
			it.Equal(client.EmitterCore.Config.PollFrequency, 500*time.Millisecond),
			it.Equal(client.EmitterCore.Config.PollerPool, 1),
		)
		client.Close()
	})

	t.Run("Method chaining maintains fluent API", func(t *testing.T) {
		mock := &mockEnqueue{}

		// Should be able to chain methods in any order
		client1, err1 := sqs.Endpoint().
			WithBatchSize(5).
			WithKernel(swarm.WithAgent("test1")).
			WithService(mock).
			Build("test")

		client2, err2 := sqs.Endpoint().
			WithService(mock).
			WithKernel(swarm.WithAgent("test2")).
			WithBatchSize(5).
			Build("test")

		it.Then(t).Should(
			it.Nil(err1),
			it.Nil(err2),
			it.Equal(client1.EmitterCore.Config.Agent, "test1"),
			it.Equal(client2.EmitterCore.Config.Agent, "test2"),
		)

		client1.Close()
		client2.Close()
	})

	t.Run("Multiple kernel options", func(t *testing.T) {
		mock := &mockEnqueue{}
		enqueuer, err := sqs.Emitter().
			WithKernel(swarm.WithAgent("service1")).
			WithKernel(swarm.WithPollFrequency(100 * time.Millisecond)).
			WithService(mock).
			Build("test")

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(enqueuer.Config.Agent, "service1"),
			it.Equal(enqueuer.Config.PollFrequency, 100*time.Millisecond),
		)
		enqueuer.Close()
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
