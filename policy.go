//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"time"

	"github.com/fogfish/swarm/backoff"
)

/*

Policy defines behavior of queue
*/
type Policy struct {
	BackoffIO     backoff.Seq
	PollFrequency time.Duration
	TimeToFlight  time.Duration
	QueueCapacity int
}

func NewPolicy() *Policy { return DefaultPolicy() }

/*

WithBackoffIO configures retry of queue I/O

  swarm.NewPolicy().
    WithBackoffIO(backoff.Exp(...))
*/
func (p *Policy) WithBackoffIO(policy backoff.Seq) *Policy {
	p.BackoffIO = policy
	return p
}

/*

WithPollFrequency configures frequency of polling loop

  swarm.NewPolicy().
    WithPollFrequency(5*time.Seconds)
*/
func (p *Policy) WithPollFrequency(t time.Duration) *Policy {
	p.PollFrequency = t
	return p
}

func (p *Policy) WithTimeToFlight(t time.Duration) *Policy {
	p.TimeToFlight = t
	return p
}

func (p *Policy) WithQueueCapacity(capacity int) *Policy {
	p.QueueCapacity = capacity
	return p
}

// default policy for the queue
func DefaultPolicy() *Policy {
	return &Policy{
		/*
		 by default all i/o is shield by exponential backoff retries.
		*/
		BackoffIO: backoff.Exp(10*time.Millisecond, 10, 0.5),

		/*
		 frequency to poll queueing api
		*/
		PollFrequency: 10 * time.Millisecond,

		/*
		 in-flight message timeout
		*/
		TimeToFlight: 5 * time.Second,

		// TBD ...
		QueueCapacity: 100,
	}
}
