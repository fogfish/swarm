//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package websocket

import (
	"time"

	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
)

// Environment variable to config event source
const EnvConfigEventType = "CONFIG_SWARM_WS_EVENT_TYPE"
const EnvConfigSourceWebSocket = "CONFIG_SWARM_WS_URL"

type Option = opts.Option[Client]

var (
	// Passes AWS SQS client instance to broker
	WithService = opts.ForType[Client, Gateway]()
)

// Global "kernel" configuration.
func WithConfig(opt ...opts.Option[swarm.Config]) Option {
	return opts.Type[Client](func(c *Client) error {
		config := swarm.NewConfig()
		if err := opts.Apply(&config, opt); err != nil {
			return err
		}

		// Mandatory overrides
		config.PollFrequency = 5 * time.Microsecond

		c.config = config
		return nil
	})
}

var defs = []Option{WithConfig()}
