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
	"fmt"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/logger/v3"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
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

// controls yield time before kernel is closed
const yield_before_close = 5 * time.Millisecond

func TestBridge(t *testing.T) {
	codec := encoding.NewCodecJson[string]()
	config := swarm.Config{PollFrequency: 0 * time.Millisecond}

	//
	mockit := func(n int) (*Dequeuer, *bridge) {
		seq := []swarm.Bag{}
		for i := 0; i < n; i++ {
			val := strconv.Itoa(i + 1)
			seq = append(seq,
				swarm.Bag{
					Ctx:    &swarm.Context{Category: "test", Digest: val},
					Object: []byte(fmt.Sprintf(`"%s"`, val)), // JSON is expected
				},
			)
		}

		brdg := mockBridge(seq)
		k := NewDequeuer(brdg, config)
		go func() {
			time.Sleep(yield_before_close)
			k.Close()
		}()

		return k, brdg
	}

	t.Run("None", func(t *testing.T) {
		k, _ := mockit(1)
		Dequeue(k, "test", codec)
		k.Await()
	})

	t.Run("Dequeue.1", func(t *testing.T) {
		k, brdg := mockit(1)
		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() { ack <- <-rcv }()
		k.Await()

		it.Then(t).Should(
			it.Seq(brdg.ack).Equal(`1`),
		)
	})

	t.Run("Dequeue.N", func(t *testing.T) {
		k, brdg := mockit(3)
		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			ack <- <-rcv
			ack <- <-rcv
		}()
		k.Await()

		it.Then(t).Should(
			it.Seq(brdg.ack).Equal(`1`, `2`, `3`),
		)
	})

	t.Run("Error.1", func(t *testing.T) {
		k, brdg := mockit(1)
		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			x := <-rcv
			ack <- x.Fail(fmt.Errorf("failed"))
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(brdg.Status).Contain("failed"),
		)
	})

	t.Run("Error.N.1", func(t *testing.T) {
		k, brdg := mockit(3)
		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			x := <-rcv
			ack <- x.Fail(fmt.Errorf("failed"))
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(brdg.Status).Contain("failed"),
		)
	})

	t.Run("Error.N.2", func(t *testing.T) {
		k, brdg := mockit(3)
		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			x := <-rcv
			ack <- x.Fail(fmt.Errorf("failed"))
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(brdg.Status).Contain("failed"),
		)
	})

	t.Run("Error.N.3", func(t *testing.T) {
		k, brdg := mockit(3)
		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			ack <- <-rcv
			x := <-rcv
			ack <- x.Fail(fmt.Errorf("failed"))
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(brdg.Status).Contain("failed"),
		)
	})

	t.Run("Timeout.1", func(t *testing.T) {
		k, brdg := mockit(1)
		rcv, _ := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() { <-rcv }()
		k.Await()

		it.Then(t).Should(
			it.Fail(brdg.Status).Contain("timeout"),
		)
	})

	t.Run("Timeout.N.1", func(t *testing.T) {
		k, brdg := mockit(3)
		rcv, _ := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			<-rcv
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(brdg.Status).Contain("timeout"),
		)
	})

	t.Run("Timeout.N.2", func(t *testing.T) {
		k, brdg := mockit(3)
		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			<-rcv
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(brdg.Status).Contain("timeout"),
		)
	})

	t.Run("Timeout.N.3", func(t *testing.T) {
		k, brdg := mockit(3)
		rcv, ack := Dequeue(k, "test", codec)

		// Note: in real apps receive loop is always go function
		go func() {
			ack <- <-rcv
			ack <- <-rcv
			<-rcv
		}()
		k.Await()

		it.Then(t).Should(
			it.Fail(brdg.Status).Contain("timeout"),
		)
	})
}

//------------------------------------------------------------------------------

// bridge mock
type bridge struct {
	*Bridge
	seq []swarm.Bag
	ack []string
	err error
}

func mockBridge(seq []swarm.Bag) *bridge {
	return &bridge{
		Bridge: NewBridge(2 * time.Millisecond),
		seq:    seq,
	}
}

func (s *bridge) Ack(ctx context.Context, digest string) error {
	if err := s.Bridge.Ack(ctx, digest); err != nil {
		return err
	}

	s.ack = append(s.ack, digest)
	return nil
}

func (s *bridge) Run() {
	s.err = s.Bridge.Dispatch(s.seq)
}

// Note: simplify assertion
func (s *bridge) Status() error {
	// Note: due to faked "handler" there is raise on setting s.err
	//       in Lambda the Dispatch returns value directly to lambda handler
	if s.err == nil {
		time.Sleep(10 * yield_before_close)
	}
	return s.err
}
