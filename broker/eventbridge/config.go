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

	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
)

// Environment variable to config event source
const EnvConfigSourceEventBridge = "CONFIG_SWARM_SOURCE_EVENTBRIDGE"

// Environment variable to config event agent
const EnvConfigEventAgent = "CONFIG_SWARM_EVENT_AGENT"

type Option = opts.Option[Client]

var defs = []Option{
	WithConfig(
		swarm.WithSource(os.Getenv(EnvConfigEventAgent)),
		swarm.WithLogStdErr(),
		swarm.WithConfigFromEnv(),
	),
}

var (
	// Configures
	WithService = opts.ForType[Client, EventBridge]()

	// Explicitly configures the event bus to use, (by default it uses environment variable CONFIG_SWARM_SOURCE_EVENTBRIDGE)
	WithEventBus = opts.ForName[Client, string]("bus")
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
