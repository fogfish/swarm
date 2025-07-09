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
	"time"

	"github.com/fogfish/swarm"
)

// Bridge Lambda's main function to [Cathode] interface
// Bridge is single threaded and should be used in the context of Lambda handler only.
type Bridge struct {
	timeToFlight time.Duration
	inflight     map[string]struct{}

	// I/O channels
	inputCh chan []swarm.Bag
	inputCx context.Context
	replyCh chan error

	// Control-plane ctrlPreempt emitter loop (MUST be non-blocking)
	ctrlPreempt chan chan struct{}
}

func NewBridge(cfg swarm.Config) *Bridge {
	return builder().Bridge(cfg)
}

func newBridge(cfg swarm.Config) *Bridge {
	return &Bridge{
		inputCh:      make(chan []swarm.Bag),
		replyCh:      make(chan error),
		timeToFlight: cfg.TimeToFlight,
	}
}

// Dispatch the batch of messages in the context of Lambda handler.
//
//	lambda.Start(
//		func(evt events.CloudWatchEvent) error {
//			...
//			bridge.Dispatch(bag)
//		}
//	)
func (s *Bridge) Dispatch(ctx context.Context, seq []swarm.Bag) error {
	s.inflight = map[string]struct{}{}
	for _, bag := range seq {
		s.inflight[bag.Digest] = struct{}{}
	}

	reqctx, cancel := context.WithTimeout(context.Background(), s.timeToFlight)
	defer cancel()

	s.inputCx = reqctx
	s.inputCh <- seq

	select {
	case err := <-s.replyCh:
		if s.ctrlPreempt != nil {
			sack := make(chan struct{})
			s.ctrlPreempt <- sack
			select {
			case <-sack:
			case <-reqctx.Done():
				return swarm.ErrTimeout("ack", s.timeToFlight)
			}
			close(sack)
		}

		return err
	case <-reqctx.Done():
		if s.ctrlPreempt != nil {
			sack := make(chan struct{})
			s.ctrlPreempt <- sack
			select {
			case <-sack:
			case <-ctx.Done():
			}
			close(sack)
		}

		return swarm.ErrTimeout("ack", s.timeToFlight)

	case <-ctx.Done():
		return nil
	}
}

// Ask converts input of Lambda handler to the context of the kernel
func (s *Bridge) Ask(ctx context.Context) ([]swarm.Bag, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case bag := <-s.inputCh:
		return bag, nil
	}
}

// Acknowledge processed message, allowing lambda handler progress
func (s *Bridge) Ack(ctx context.Context, digest string) error {
	delete(s.inflight, digest)
	if len(s.inflight) == 0 {
		select {
		case <-ctx.Done():
			return nil
		case <-s.inputCx.Done():
			return nil
		case s.replyCh <- nil:
			return nil
		}
	}
	return nil
}

// Acknowledge error, allowing lambda handler progress
func (s *Bridge) Err(ctx context.Context, digest string, err error) error {
	delete(s.inflight, digest)
	select {
	case <-ctx.Done():
		return nil
	case <-s.inputCx.Done():
		return nil
	case s.replyCh <- err:
		return nil
	}
}
