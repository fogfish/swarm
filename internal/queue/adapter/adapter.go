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
	logger logger.Logger
}

func New(policy *swarm.Policy, logger logger.Logger) *Adapter {
	return &Adapter{
		Policy: policy,
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
func Enq(q *Adapter, f func(*swarm.Bag) error) chan *swarm.BagStdErr {
	q.logger.Info("init enqueue adapter")

	sock := make(chan *swarm.BagStdErr, q.Policy.QueueCapacity)

	pipe.ForEach(sock, func(bag *swarm.BagStdErr) {
		// TODO: add message type control
		// the control message to ensure that channels are free
		// raw := msg.Object.Bytes()
		if string(bag.Object[:3]) == "+++" {
			bag.StdErr(nil)
			q.logger.Info("end of stream")
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
func Deq(q *Adapter, f func() (*swarm.Bag, error)) chan *swarm.Bag {
	q.logger.Info("init dequeue adapter")

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
func Ack(q *Adapter, f func(*swarm.Bag) error) chan *swarm.Bag {
	q.logger.Info("init ack adapter")

	conf := make(chan *swarm.Bag)

	pipe.ForEach(conf, func(bag *swarm.Bag) {
		err := q.Policy.BackoffIO.Retry(func() error { return f(bag) })
		if err != nil {
			q.logger.Error("Unable to ack message %v", err)
		}
	})

	return conf
}
