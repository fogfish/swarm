//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package adapter

import (
	"fmt"

	"github.com/fogfish/golem/pipe"
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
func Send(q *Adapter, f func(*swarm.Bag) error) chan *swarm.BagStdErr {
	q.logger.Notice("init send")

	sock := make(chan *swarm.BagStdErr, q.Policy.QueueCapacity)

	pipe.ForEach(sock, func(bag *swarm.BagStdErr) {
		// TODO: add message type control
		// the control message to ensure that channels are free
		// raw := msg.Object.Bytes()
		if string(bag.Object[:3]) == "+++" {
			bag.StdErr(nil)
			return
		}

		err := q.Policy.BackoffIO.Retry(func() error { return f(&bag.Bag) })
		if err != nil {
			// TODO: the progress of sender is blocked until
			//       failed message is consumed
			bag.StdErr(err)
			q.logger.Debug("failed to send message %v", err)
		}
	})

	return sock
}

/*

Recv create go routine to adapt async i/o over Golang channel to synchronous
calls of queueing system interface
*/
func Recv(q *Adapter, f func() (*swarm.Bag, error)) chan *swarm.Bag {
	q.logger.Notice("init recv")

	return pipe.From(0, q.Policy.PollFrequency, func() (*swarm.Bag, error) {
		var msg *swarm.Bag
		err := q.Policy.BackoffIO.Retry(
			func() (e error) {
				msg, e = f()
				return
			},
		)
		if err != nil {
			q.logger.Error("Unable to receive message %v", err)
			return nil, err
		}
		if msg == nil {
			return nil, fmt.Errorf("Nothing is received")
		}

		return msg, nil
	})
}

/*

Conf create go routine to adapt async i/o over Golang channel to synchronous
calls of queueing system interface
*/
func Conf(q *Adapter, f func(*swarm.Bag) error) chan *swarm.Bag {
	q.logger.Notice("init conf")

	conf := make(chan *swarm.Bag)

	pipe.ForEach(conf, func(bag *swarm.Bag) {
		// switch msg := bag.Object.(type) {
		// case *swarm.Msg:
		err := q.Policy.BackoffIO.Retry(func() error { return f(bag) })
		if err != nil {
			q.logger.Error("Unable to conf message %v", err)
		}
		// default:
		// q.logger.Notice("Unsupported conf type %v", bag)
		// }
	})

	return conf
}
