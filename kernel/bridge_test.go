//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/logger/v3"
	"github.com/fogfish/swarm"
)

func init() {
	slog.SetDefault(
		logger.New(
			logger.WithSourceShorten(),
			logger.WithoutTimestamp(),
			logger.WithLogLevel(slog.LevelDebug),
		),
	)
}

func TestBridge(t *testing.T) {

	seq := []swarm.Bag{
		{Ctx: &swarm.Context{Category: "test", Digest: "1"}, Object: []byte(`"1"`)},
		{Ctx: &swarm.Context{Category: "test", Digest: "2"}, Object: []byte(`"2"`)},
	}

	t.Run("None", func(t *testing.T) {
		k := NewDequeuer(mockSpawner(seq), swarm.Config{PollFrequency: 1 * time.Second})
		go k.Await()
		k.Close()
	})

	t.Run("Dequeue.1", func(t *testing.T) {
		mock := mockSpawner(seq)
		k := NewDequeuer(mock, swarm.Config{PollFrequency: 10 * time.Millisecond})
		rcv, ack := Dequeue(k, "test", swarm.NewCodecJson[string]())
		go k.Await()

		ack <- <-rcv
		ack <- <-rcv

		k.Close()

		it.Then(t).Should(
			it.Seq(mock.ack).Equal(`1`, `2`),
		)
	})

	t.Run("Timeout", func(t *testing.T) {
		mock := mockSpawner(seq)
		k := NewDequeuer(mock, swarm.Config{PollFrequency: 0 * time.Millisecond})
		rcv, ack := Dequeue(k, "test", swarm.NewCodecJson[string]())
		go k.Await()

		ack <- <-rcv
		<-rcv
		time.Sleep(1 * time.Millisecond)

		k.Close()
		time.Sleep(1 * time.Millisecond)

		it.Then(t).ShouldNot(
			it.Nil(mock.err),
		)
	})

}

//------------------------------------------------------------------------------

type spawner struct {
	*Bridge
	seq []swarm.Bag
	ack []string
	err error
}

func mockSpawner(seq []swarm.Bag) *spawner {
	return &spawner{
		Bridge: NewBridge(1 * time.Millisecond),
		seq:    seq,
	}
}

func (s *spawner) Ack(ctx context.Context, digest string) error {
	if err := s.Bridge.Ack(ctx, digest); err != nil {
		return err
	}

	s.ack = append(s.ack, digest)
	return nil
}

func (s *spawner) Run() {
	s.err = s.Bridge.Dispatch(s.seq)
}
