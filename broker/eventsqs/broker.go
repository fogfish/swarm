//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
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
	"github.com/fogfish/swarm/internal/router"
)

// New creates broker for AWS EventBridge
func New(queue string, opts ...swarm.Option) (swarm.Broker, error) {
	conf := swarm.NewConfig()
	for _, opt := range opts {
		opt(&conf)
	}

	bro, err := sqs.New(queue, opts...)
	if err != nil {
		return nil, err
	}

	return &broker{
		Broker: bro,
		config: conf,
		router: router.New(nil),
	}, nil
}

type broker struct {
	swarm.Broker
	config swarm.Config
	router *router.Router
}

func (b *broker) Dequeue(category string, channel swarm.Channel) swarm.Dequeue {
	b.Broker.Dequeue(category, channel)
	b.router.Register(category, b.config.DequeueCapacity)

	return b.router
}

func (b *broker) Await() {
	starter := lambda.Start

	type Mock interface{ Start(interface{}) }
	if b.config.Service != nil {
		service, ok := b.config.Service.(Mock)
		if ok {
			starter = service.Start
		}
	}

	starter(
		func(events events.SQSEvent) error {
			for _, evt := range events.Records {
				bag := swarm.Bag{
					Category: attr(&evt, "Category"),
					Object:   []byte(evt.Body),
					Digest:   evt.ReceiptHandle,
				}
				if err := b.router.Dispatch(bag); err != nil {
					return err
				}
			}

			return b.router.Await(b.config.TimeToFlight)
		},
	)
}

func attr(msg *events.SQSMessage, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}
