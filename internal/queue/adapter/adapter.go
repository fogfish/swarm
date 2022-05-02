//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package adapter

import (
	"context"
	"time"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

/*

Adapter to queueing systems implement utility function to convert pure
api calls into channel messaging
*/
type Adapter struct {
	Policy *swarm.Policy

	sys    swarm.System
	logger logger.Logger
}

func New(sys swarm.System, policy *swarm.Policy, logger logger.Logger) *Adapter {
	return &Adapter{
		Policy: policy,

		sys:    sys,
		logger: logger,
	}
}

/*

Send builds a channel to listen for incoming Bags and relay it to the function
*/
/*

Send create go routine to adapt async i/o over Golang channel to synchronous
calls of queueing system interface

AdaptSend builds a channel to listen for incoming Bags and relay it to the function
*/
func Send(q *Adapter, f func(msg *swarm.Bag) error) chan<- *swarm.Bag {
	sock := make(chan *swarm.Bag, q.Policy.QueueCapacity)

	q.sys.Go(func(ctx context.Context) {
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
				// TODO: add message type control
				// the control message to ensure that channels are free
				raw := msg.Object.Bytes()
				if string(raw[:3]) == "+++" {
					msg.StdErr <- msg.Object
					continue
				}

				err := q.Policy.BackoffIO.Retry(func() error { return f(msg) })
				if err != nil {
					// TODO: the progress of sender is blocked until
					//       failed message is consumed
					msg.StdErr <- msg.Object
					q.logger.Debug("failed to send message %v", err)
				}
			}
		}
	})

	return sock
}

/*

Recv create go routine to adapt async i/o over Golang channel to synchronous
calls of queueing system interface
*/
func Recv(q *Adapter, f func() (*swarm.Bag, error)) <-chan *swarm.Bag {
	freq := q.Policy.PollFrequency
	sock := make(chan *swarm.Bag)

	q.sys.Go(func(ctx context.Context) {
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
				var msg *swarm.Bag
				err := q.Policy.BackoffIO.Retry(
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

Conf create go routine to adapt async i/o over Golang channel to synchronous
calls of queueing system interface
*/
func Conf(q *Adapter, f func(msg *swarm.Msg) error) chan<- *swarm.Bag {
	conf := make(chan *swarm.Bag)

	q.sys.Go(func(ctx context.Context) {
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
				case *swarm.Msg:
					err := q.Policy.BackoffIO.Retry(func() error { return f(msg) })
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
