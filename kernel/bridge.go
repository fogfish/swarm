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
type Bridge struct {
	timeToFlight time.Duration
	inflight     map[string]struct{}
	session      chan error
	ch           chan []swarm.Bag
}

func NewBridge(timeToFlight time.Duration) *Bridge {
	return &Bridge{
		ch:           make(chan []swarm.Bag),
		session:      make(chan error),
		timeToFlight: timeToFlight,
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
func (s *Bridge) Dispatch(seq []swarm.Bag) error {
	s.inflight = map[string]struct{}{}
	for _, bag := range seq {
		s.inflight[bag.Digest] = struct{}{}
	}

	s.ch <- seq

	select {
	case err := <-s.session:
		return err
	case <-time.After(s.timeToFlight):
		return swarm.ErrTimeout("ack", s.timeToFlight)
	}
}

// Ask converts input of Lambda handler to the context of the kernel
func (s *Bridge) Ask(ctx context.Context) ([]swarm.Bag, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case bag := <-s.ch:
		return bag, nil
	}
}

// Acknowledge processed message, allowing lambda handler progress
func (s *Bridge) Ack(ctx context.Context, digest string) error {
	delete(s.inflight, digest)
	if len(s.inflight) == 0 {
		s.session <- nil
	}

	return nil
}

// Acknowledge error, allowing lambda handler progress
func (s *Bridge) Err(ctx context.Context, digest string, err error) error {
	delete(s.inflight, digest)
	s.session <- err
	return nil
}
