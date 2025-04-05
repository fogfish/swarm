//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"time"

	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
)

// Environment variable to config event source
const EnvConfigSourceEventBridge = "CONFIG_SWARM_SOURCE_EVENTBRIDGE"

type Option = opts.Option[Client]

var defs = []Option{WithConfig()}

var (
	// Configures
	WithService = opts.ForType[Client, EventBridge]()
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

// type Option func(*Client)

// var defs = []Option{WithConfig()}

// func WithConfig(opts ...swarm.Option) Option {
// 	return func(c *Client) {
// 		config := swarm.NewConfig()
// 		for _, opt := range opts {
// 			opt(&config)
// 		}

// 		// Mandatory overrides
// 		config.PollFrequency = 5 * time.Microsecond
// 		config.PacketCodec = encoding.ForBytesJB64()

// 		c.config = config
// 	}
// }

// func WithService(service EventBridge) Option {
// 	return func(c *Client) {
// 		c.service = service
// 	}
// }
