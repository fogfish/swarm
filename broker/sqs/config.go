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

// Environment variable to config event source
const EnvConfigSourceSQS = "CONFIG_SWARM_SOURCE_SQS"

type Option func(*Client)

var defs = []Option{WithConfig()}

// Global "kernel" configuration.
func WithConfig(opts ...swarm.Option) Option {
	return func(c *Client) {
		config := swarm.NewConfig()
		for _, opt := range opts {
			opt(&config)
		}

		if c.batchSize == 0 {
			c.batchSize = 1
		}

		c.config = config
	}
}

// Passes AWS SQS client instance to broker
func WithService(service SQS) Option {
	return func(c *Client) {
		c.service = service
	}
}

// Configure's SQS batch size.
// Note: AWS SQS limits the MaxNumberOfMessages to 10 per poller.
// This config increases number of poller
func WithBatchSize(batch int) Option {
	return func(c *Client) {
		if batch <= 10 {
			c.batchSize = batch
			return
		}

		const maxNumberOfMessages = 10
		c.batchSize = maxNumberOfMessages
		c.config.PollerPool = batch/maxNumberOfMessages + 1
	}
}
