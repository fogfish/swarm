//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"os"
	"time"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
)

type Option func(*Client)

var defs = []Option{WithConfig(), WithEnv()}

func WithConfig(opts ...swarm.Option) Option {
	return func(c *Client) {
		config := swarm.NewConfig()
		for _, opt := range opts {
			opt(&config)
		}

		// Mandatory overrides
		config.PollFrequency = 5 * time.Microsecond
		config.Codec = encoding.NewCodecPacket()

		c.config = config
	}
}

func WithService(service EventBridge) Option {
	return func(c *Client) {
		c.service = service
	}
}

const EnvEventBridge = "CONFIG_SWARM_EVENT_BRIDGE"

func WithEnv() Option {
	return func(c *Client) {
		if val, has := os.LookupEnv(EnvEventBridge); has {
			c.bus = val
		}
	}
}
