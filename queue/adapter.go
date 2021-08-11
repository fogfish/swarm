//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package queue

import (
	"context"
	"time"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/backoff"
)

/*

Utility methods to implement integrations with queueing systems

*/

/*

queue system policy
*/
type tPolicy struct {
	io            backoff.Seq
	frequencyPoll time.Duration
}

func defaultPolicy() *tPolicy {
	return &tPolicy{
		/*
		 by default all i/o is shield by exponential backoff retries.
		*/
		io: backoff.Exp(10*time.Millisecond, 10, 0.5),

		/*
		 frequency to poll queueing api
		*/
		frequencyPoll: 10 * time.Millisecond,
	}
}

/*

Adapter to queueing systems
*/
type Adapter struct {
	System swarm.System
	ID     string

	policy *tPolicy
	logger logger.Logger
}

/*

Adapt creates helper utility to adapt i/o with external queuing system
*/
func Adapt(sys swarm.System, service string, id string) *Adapter {
	return &Adapter{
		System: sys,
		ID:     id,

		policy: defaultPolicy(),
		logger: logger.With(logger.Note{
			"type": service,
			"q":    id,
		}),
	}
}

/*

SetPolicyIO defines policy for queue I/O operations.
*/
func (q *Adapter) SetPolicyIO(policy backoff.Seq) {
	q.policy.io = policy
}

/*

SetPolicyPoll defines frequency of queue I/O operations
*/
func (q *Adapter) SetPolicyPoll(frequency time.Duration) {
	q.policy.frequencyPoll = frequency
}

/*

SendIO create go routine to adapt async i/o over Golang channel to synchronous
calls of queueing system interface
*/
func (q *Adapter) SendIO(f func(msg *Bag) error) chan<- *Bag {
	sock := make(chan *Bag)

	q.System.Go(func(ctx context.Context) {
		q.logger.Notice("init send")
		defer close(sock)

		for {
			select {
			//
			case <-ctx.Done():
				q.logger.Notice("free send")
				return

			//
			case msg := <-sock:
				err := q.policy.io.Retry(func() error { return f(msg) })
				if err != nil {
					msg.StdErr <- msg.Object
					q.logger.Debug("failed to send message %v", err)
				}
			}
		}
	})

	return sock
}

/*

RecvIO create go routine to adapt async i/o over Golang channel to synchronous
calls of queueing system interface
*/
func (q *Adapter) RecvIO(f func() (*Bag, error)) <-chan *Bag {
	freq := q.policy.frequencyPoll
	sock := make(chan *Bag)

	q.System.Go(func(ctx context.Context) {
		q.logger.Notice("init recv")
		defer close(sock)

		for {
			select {
			//
			case <-ctx.Done():
				q.logger.Notice("free recv")
				return

			//
			case <-time.After(freq):
				var msg *Bag
				err := q.policy.io.Retry(
					func() (e error) {
						msg, e = f()
						return
					},
				)

				if err != nil {
					q.logger.Error("Unable to receive message %v", err)
					break
				}

				if msg != nil {
					// TODO: send is blocked here until actor consumes the message.
					//       https://go101.org/article/channel-closing.html
					sock <- msg
				}
			}
		}
	})

	return sock
}

/*

ConfIO create go routine to adapt async i/o over Golang channel to synchronous
calls of queueing system interface
*/
func (q *Adapter) ConfIO(f func(msg *Msg) error) chan<- *Bag {
	conf := make(chan *Bag)

	q.System.Go(func(ctx context.Context) {
		q.logger.Notice("init conf")
		defer close(conf)

		for {
			select {
			//
			case <-ctx.Done():
				q.logger.Notice("free conf")
				return

			//
			case bag := <-conf:
				switch msg := bag.Object.(type) {
				case *Msg:
					err := q.policy.io.Retry(func() error { return f(msg) })
					if err != nil {
						q.logger.Error("Unable to conf message %v", err)
					}
				default:
					q.logger.Notice("Unsupported conf type %v", bag)
				}
			}
		}
	})

	return conf
}
