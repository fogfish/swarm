//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventddb

import (
	"time"

	"github.com/fogfish/logger/x/xlog"
	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// Environment define name of source DynamoDB table
const EnvConfigSourceDynamoDB = "CONFIG_SWARM_SOURCE_DDB"

// Environment define name of target DynamoDB table
const EnvConfigTargetDynamoDB = "CONFIG_SWARM_TARGET_DDB"

func Must[T any](v T, err error) T {
	if err != nil {
		xlog.Emergency("eventddb broker has failed", err)
	}
	return v
}

type ListenerBuilder struct{ *builder[*ListenerBuilder] }

func Listener() *ListenerBuilder {
	b := &ListenerBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *ListenerBuilder) Build() (*kernel.ListenerIO, error) {
	client, err := b.build()
	if err != nil {
		return nil, err
	}

	bridge := &bridge{kernel.NewBridge(client.config)}
	return kernel.NewListener(bridge, client.config), nil
}

//------------------------------------------------------------------------------

type builder[T any] struct {
	b          T
	kernelOpts []opts.Option[swarm.Config]
}

// newBuilder creates new builder for EventDDB broker configuration.
func newBuilder[T any](b T) *builder[T] {
	kopts := []opts.Option[swarm.Config]{
		swarm.WithLogStdErr(),
		swarm.WithConfigFromEnv(),
	}

	return &builder[T]{
		b:          b,
		kernelOpts: kopts,
	}
}

// WithKernel configures swarm kernel options for advanced usage.
func (b *builder[T]) WithKernel(opts ...opts.Option[swarm.Config]) T {
	b.kernelOpts = append(b.kernelOpts, opts...)
	return b.b
}

// build constructs the EventDDB client with configuration
func (b *builder[T]) build() (*Client, error) {
	client := &Client{
		config: swarm.NewConfig(),
	}

	if err := opts.Apply(&client.config, b.kernelOpts); err != nil {
		return nil, err
	}

	// Apply mandatory overrides for DynamoDB Events
	client.config.PollFrequency = 5 * time.Microsecond

	return client, nil
}
