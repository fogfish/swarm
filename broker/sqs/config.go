//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package sqs

import (
	"github.com/fogfish/swarm"
)

type Option func(*Client)

var defs = []Option{WithConfig()}

func WithConfig(opts ...swarm.Option) Option {
	return func(c *Client) {
		config := swarm.NewConfig()
		for _, opt := range opts {
			opt(&config)
		}

		c.batchSize = 1

		c.config = config
	}
}

func WithService(service SQS) Option {
	return func(c *Client) {
		c.service = service
	}
}

func WithBatchSize(batch int) Option {
	return func(c *Client) {
		c.batchSize = batch
	}
}
