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
	cfg := swarm.NewConfig()
	cfg.TimeToFlight = 100 * time.Millisecond
	bridge := &bridge{kernel.NewBridge(cfg)}

	t.Run("New", func(t *testing.T) {
		q, err := Listener().Build()
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

		err := bridge.run(context.Background(),
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
		dequeuer, err := Listener().Build()
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(dequeuer.Config.PollFrequency, 5*time.Microsecond), // Events3 override
		)
		dequeuer.Close()
	})

	t.Run("Kernel configuration", func(t *testing.T) {
		dequeuer, err := Listener().
			WithKernel(
				swarm.WithAgent("s3-processor"),
				swarm.WithTimeToFlight(30*time.Second),
			).
			Build()

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(dequeuer.Config.Agent, "s3-processor"),
			it.Equal(dequeuer.Config.TimeToFlight, 30*time.Second),
			it.Equal(dequeuer.Config.PollFrequency, 5*time.Microsecond), // Events3 override
		)
		dequeuer.Close()
	})

	t.Run("Multiple kernel options", func(t *testing.T) {
		dequeuer, err := Listener().
			WithKernel(swarm.WithAgent("service1")).
			WithKernel(swarm.WithTimeToFlight(60 * time.Second)).
			Build()

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(dequeuer.Config.Agent, "service1"),
			it.Equal(dequeuer.Config.TimeToFlight, 60*time.Second),
		)
		dequeuer.Close()
	})
}
