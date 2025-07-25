//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

func TestBuilder(t *testing.T) {
	t.Run("Listener", func(t *testing.T) {
		q, err := Listener().Build()

		it.Then(t).Should(
			it.Nil(err),
		).ShouldNot(
			it.Nil(q),
		)

		q.Close()
	})

	t.Run("ListenerWithKernel", func(t *testing.T) {
		q, err := Listener().
			WithKernel(swarm.WithLogStdErr()).
			Build()

		it.Then(t).Should(
			it.Nil(err),
		).ShouldNot(
			it.Nil(q),
		)

		q.Close()
	})

	t.Run("MustListener", func(t *testing.T) {
		q := Must(Listener().Build())

		it.Then(t).ShouldNot(
			it.Nil(q),
		)

		q.Close()
	})

}

func TestReader(t *testing.T) {
	var bag []swarm.Bag
	cfg := swarm.NewConfig()
	cfg.TimeToFlight = 100 * time.Millisecond
	bridge := &bridge{kernel.NewBridge(cfg)}

	// Test new builder pattern
	t.Run("Builder.NewDequeuer", func(t *testing.T) {
		q, err := Listener().Build()
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Builder.WithKernel", func(t *testing.T) {
		q, err := Listener().
			WithKernel(swarm.WithLogStdErr()).
			Build()
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

		err := bridge.run(context.Background(),
			events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId:     "abc-def",
						ReceiptHandle: "receipt",
						Body:          `{"sut":"test"}`,
						MessageAttributes: map[string]events.SQSMessageAttribute{
							"Category": {StringValue: aws.String("cat")},
						},
					},
				},
			},
		)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(len(bag), 1),
			it.Equal(bag[0].Category, "cat"),
			it.Equal(bag[0].Digest, "receipt"),
			it.Equiv(bag[0].Object, []byte(`{"sut":"test"}`)),
		)
	})

	t.Run("Dequeue.Timeout", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
		}()

		err := bridge.run(context.Background(),
			events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId:     "abc-def",
						ReceiptHandle: "receipt",
						Body:          `{"sut":"test"}`,
						MessageAttributes: map[string]events.SQSMessageAttribute{
							"Category": {StringValue: aws.String("cat")},
						},
					},
				},
			},
		)

		it.Then(t).ShouldNot(
			it.Nil(err),
		)
	})
}
