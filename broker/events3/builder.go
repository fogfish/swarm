//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events3

import (
	"os"
	"time"

	"github.com/fogfish/logger/x/xlog"
	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// Environment variable to config event source
const EnvConfigSourceS3 = "CONFIG_SWARM_SOURCE_S3"

// Builder provides API for configuring Events3 broker
type Builder struct {
	kernelOpts []opts.Option[swarm.Config]
}

// Channels creates new builder for Events3 broker configuration.
func Channels() *Builder {
	kopts := []opts.Option[swarm.Config]{
		swarm.WithLogStdErr(),
		swarm.WithConfigFromEnv(),
	}
	if val := os.Getenv(EnvConfigSourceS3); val != "" {
		kopts = append(kopts, swarm.WithSource(val))
	}

	return &Builder{
		kernelOpts: kopts,
	}
}

// WithKernel configures swarm kernel options for advanced usage.
func (b *Builder) WithKernel(opts ...opts.Option[swarm.Config]) *Builder {
	b.kernelOpts = append(b.kernelOpts, opts...)
	return b
}

// NewDequeuer creates dequeue routine from AWS S3 Events (read-only)
func (b *Builder) NewDequeuer() (*kernel.ListenerCore, error) {
	client, err := b.build()
	if err != nil {
		return nil, err
	}

	bridge := &bridge{kernel.NewBridge(client.config.TimeToFlight)}
	return kernel.NewListener(bridge, client.config), nil
}

func (b *Builder) MustDequeuer() *kernel.ListenerCore {
	q, err := b.NewDequeuer()
	if err != nil {
		xlog.Emergency("events3 client has failed", err)
	}
	return q
}

// build constructs the Events3 client with configuration
func (b *Builder) build() (*Client, error) {
	client := &Client{
		config: swarm.NewConfig(),
	}

	if err := opts.Apply(&client.config, b.kernelOpts); err != nil {
		return nil, err
	}

	client.config.PollFrequency = 5 * time.Microsecond

	return client, nil
}
