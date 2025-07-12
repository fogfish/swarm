//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events3

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

func TestReader(t *testing.T) {
	var bag []swarm.Bag
	bridge := &bridge{kernel.NewBridge(100 * time.Millisecond)}

	t.Run("New", func(t *testing.T) {
		q, err := Channels().NewDequeuer()
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
			S3Event{
				Records: []json.RawMessage{[]byte(`{"sut":"test"}`)},
			},
		)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(len(bag), 1),
			it.Equal(bag[0].Category, Category),
			it.Equiv(bag[0].Object, []byte(`{"sut":"test"}`)),
		)
	})

	t.Run("Dequeue.Timeout", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
		}()

		err := bridge.run(
			S3Event{
				Records: []json.RawMessage{[]byte(`{"sut":"test"}`)},
			},
		)

		it.Then(t).ShouldNot(
			it.Nil(err),
		)
	})
}

func TestBuilder(t *testing.T) {
	t.Run("Simple case with sensible defaults", func(t *testing.T) {
		dequeuer, err := Channels().NewDequeuer()
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(dequeuer.Config.PollFrequency, 5*time.Microsecond), // Events3 override
		)
		dequeuer.Close()
	})

	t.Run("Kernel configuration", func(t *testing.T) {
		dequeuer, err := Channels().
			WithKernel(
				swarm.WithSource("s3-processor"),
				swarm.WithTimeToFlight(30*time.Second),
			).
			NewDequeuer()

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(dequeuer.Config.Agent, "s3-processor"),
			it.Equal(dequeuer.Config.TimeToFlight, 30*time.Second),
			it.Equal(dequeuer.Config.PollFrequency, 5*time.Microsecond), // Events3 override
		)
		dequeuer.Close()
	})

	t.Run("Method chaining maintains fluent API", func(t *testing.T) {
		dequeuer1, err1 := Channels().
			WithKernel(swarm.WithSource("test1")).
			NewDequeuer()

		dequeuer2, err2 := Channels().
			WithKernel(swarm.WithSource("test2")).
			NewDequeuer()

		it.Then(t).Should(
			it.Nil(err1),
			it.Nil(err2),
			it.Equal(dequeuer1.Config.Agent, "test1"),
			it.Equal(dequeuer2.Config.Agent, "test2"),
		)

		dequeuer1.Close()
		dequeuer2.Close()
	})

	t.Run("Multiple kernel options", func(t *testing.T) {
		dequeuer, err := Channels().
			WithKernel(swarm.WithSource("service1")).
			WithKernel(swarm.WithTimeToFlight(60 * time.Second)).
			NewDequeuer()

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(dequeuer.Config.Agent, "service1"),
			it.Equal(dequeuer.Config.TimeToFlight, 60*time.Second),
		)
		dequeuer.Close()
	})
}

func TestBuilderBackwardCompatibility(t *testing.T) {
	// Verify that old API still works during transition period
	t.Run("Legacy NewDequeuer still works", func(t *testing.T) {
		q, err := Channels().NewDequeuer()

		it.Then(t).Should(it.Nil(err))
		q.Close()
	})
}
