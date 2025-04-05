//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package sqs

import (
	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
)

// Environment variable to config event source
const EnvConfigSourceSQS = "CONFIG_SWARM_SOURCE_SQS"

type Option = opts.Option[Client]

var (
	// Passes AWS SQS client instance to broker
	WithService = opts.ForType[Client, SQS]()

	// Configure's SQS batch size.
	// Note: AWS SQS limits the MaxNumberOfMessages to 10 per poller.
	// This config increases number of poller
	WithBatchSize = opts.ForName[Client, int]("batchSize",
		func(c *Client, batch int) error {
			if batch <= 10 {
				c.batchSize = batch
				return nil
			}

			const maxNumberOfMessages = 10
			c.batchSize = maxNumberOfMessages
			c.config.PollerPool = batch/maxNumberOfMessages + 1

			return nil
		},
	)
)

// Global "kernel" configuration.
func WithConfig(opt ...opts.Option[swarm.Config]) Option {
	return opts.Type[Client](func(c *Client) error {
		config := swarm.NewConfig()
		if err := opts.Apply(&config, opt); err != nil {
			return err
		}

		c.config = config
		return nil
	})
}

var defs = []Option{WithConfig()}
