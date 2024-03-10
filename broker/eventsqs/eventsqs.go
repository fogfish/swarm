//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/internal/kernel"
)

// New creates broker for AWS SQS (serverless events)
func New(queue string, opts ...swarm.Option) (swarm.Broker, error) {
	cli, err := sqs.NewSQS(queue, opts...)
	if err != nil {
		return nil, err
	}

	config := swarm.NewConfig()
	for _, opt := range opts {
		opt(&config)
	}

	starter := lambda.Start

	type Mock interface{ Start(interface{}) }
	if config.Service != nil {
		service, ok := config.Service.(Mock)
		if ok {
			starter = service.Start
		}
	}

	sls := spawner{f: starter, c: config}

	return kernel.New(cli, sls, config), nil
}

//------------------------------------------------------------------------------

type spawner struct {
	c swarm.Config
	f func(any)
}

func (s spawner) Spawn(k *kernel.Kernel) error {
	s.f(
		func(events events.SQSEvent) error {
			bag := make([]swarm.Bag, len(events.Records))
			for i, evt := range events.Records {
				bag[i] = swarm.Bag{
					Category: attr(&evt, "Category"),
					Object:   []byte(evt.Body),
					Digest:   swarm.Digest{Brief: evt.ReceiptHandle},
				}
			}

			return k.Dispatch(bag, s.c.TimeToFlight)
		},
	)

	return nil
}

func (s spawner) Ack(digest string) error   { return nil }
func (s spawner) Ask() ([]swarm.Bag, error) { return nil, nil }

func attr(msg *events.SQSMessage, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}
