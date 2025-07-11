//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/broadcast"
)

type config struct {
	kernel swarm.Config

	yieldBeforeClose time.Duration // timeout before kernel is closed
	timeToEmit       time.Duration // delay for emitter function
}

func newConfig() config {
	return config{
		kernel:           swarm.NewConfig(),
		yieldBeforeClose: 50 * time.Millisecond,
		timeToEmit:       10 * time.Microsecond,
	}
}

type mockFactory struct {
	// messaging kernel is running in the externally preemptable mode
	isExternallyPreemptable bool

	ctrlPreempt *broadcast.Broadcaster
}

func (b *mockFactory) enableExternalPreemption() {
	b.isExternallyPreemptable = true
	b.ctrlPreempt = broadcast.New()
}

func (b *mockFactory) Bridge(cfg config, seq []swarm.Bag) *mockBridge {
	bridge := newMockBridge(cfg, seq)
	bridge.ctrlPreempt = b.ctrlPreempt

	return bridge
}

func (b *mockFactory) EmitterCore(cfg config) *mockEmitter {
	return newMockEmitter(cfg)
}

func (b *mockFactory) ListenerCore(ack chan string, seq []swarm.Bag) *mockListener {
	return newMockCathode(ack, seq)
}

func (b *mockFactory) Emitter(emitter Emitter, cfg config) *EmitterCore {
	enqueuer := newEmitter(emitter, cfg.kernel)
	enqueuer.ctrlPreempt = b.ctrlPreempt

	go func() {
		time.Sleep(cfg.yieldBeforeClose)
		enqueuer.Close()
	}()

	return enqueuer
}

func (b *mockFactory) Listener(cathode Listener, cfg config) *ListenerCore {
	dequeuer := newListener(cathode, cfg.kernel)
	// auto close is essential for testing
	go func() {
		time.Sleep(cfg.yieldBeforeClose)
		dequeuer.Close()
	}()

	return dequeuer
}

func (b *mockFactory) Bag(n int) []swarm.Bag {
	seq := []swarm.Bag{}
	for i := range n {
		val := strconv.Itoa(i + 1)
		seq = append(seq,
			swarm.Bag{
				Category:  "test",
				Digest:    val,
				IOContext: "context",
				Object:    fmt.Appendf(nil, `"%s"`, val), // JSON is expected
			},
		)
	}
	return seq
}

//------------------------------------------------------------------------------

type mockBridge struct {
	*Bridge
	cfg       config
	seq       []swarm.Bag
	ack       []string
	err       error
	stoppedAt time.Time
}

func newMockBridge(cfg config, seq []swarm.Bag) *mockBridge {
	return &mockBridge{
		Bridge: NewBridge(cfg.kernel),
		cfg:    cfg,
		seq:    seq,
	}
}

func (s *mockBridge) Ack(ctx context.Context, digest string) error {
	if err := s.Bridge.Ack(ctx, digest); err != nil {
		return err
	}

	s.ack = append(s.ack, digest)
	return nil
}

func (s *mockBridge) Run(ctx context.Context) {
	s.err = s.Bridge.Dispatch(ctx, s.seq)
	s.stoppedAt = time.Now()
}

func (s *mockBridge) Status() error {
	// Note: due to faked "handler" there is raise on setting s.err
	//       in Lambda the Dispatch returns value directly to lambda handler
	//
	//       Only implemented to simplify assertion
	if s.err == nil {
		time.Sleep(3 * s.cfg.yieldBeforeClose)
	}
	return s.err
}

//------------------------------------------------------------------------------

type mockEmitter struct {
	cfg       config
	seq       []string
	val       chan string
	emittedAt time.Time
}

func newMockEmitter(cfg config) *mockEmitter {
	return &mockEmitter{
		cfg: cfg,
		seq: make([]string, 0),
		val: make(chan string, 1000),
	}
}

func (e *mockEmitter) Enq(ctx context.Context, bag swarm.Bag) error {
	time.Sleep(e.cfg.timeToEmit)
	e.seq = append(e.seq, string(bag.Object))
	e.emittedAt = time.Now()

	e.val <- string(bag.Object)
	return nil
}

//------------------------------------------------------------------------------

type mockListener struct {
	seq []swarm.Bag
	ack chan string
}

func newMockCathode(ack chan string, seq []swarm.Bag) *mockListener {
	return &mockListener{seq: seq, ack: ack}
}

func (c *mockListener) Ack(ctx context.Context, digest string) error {
	c.ack <- digest
	return nil
}

func (c *mockListener) Err(ctx context.Context, digest string, err error) error {
	c.ack <- digest
	return nil
}

func (c *mockListener) Ask(context.Context) ([]swarm.Bag, error) {
	return c.seq, nil
}
