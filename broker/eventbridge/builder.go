//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/fogfish/logger/x/xlog"
	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// Environment define name of source EventBus
const EnvConfigSourceEventBus = "CONFIG_SWARM_SOURCE_EVENTBUS"

// Environment define name of target EventBus
const EnvConfigTargetEventBus = "CONFIG_SWARM_TARGET_EVENTBUS"

func Must[T any](v T, err error) T {
	if err != nil {
		xlog.Emergency("eventbridge broker has failed", err)
	}
	return v
}

type EndpointBuilder struct{ *builder[*EndpointBuilder] }

func Endpoint() *EndpointBuilder {
	b := &EndpointBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *EndpointBuilder) Build(bus string) (*kernel.Kernel, error) {
	client, err := b.build(bus)
	if err != nil {
		return nil, err
	}

	bridge := &bridge{kernel.NewBridge(client.config)}
	return kernel.New(
		kernel.NewEmitter(client, client.config),
		kernel.NewListener(bridge, client.config),
	), nil
}

type EmitterBuilder struct{ *builder[*EmitterBuilder] }

func Emitter() *EmitterBuilder {
	b := &EmitterBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *EmitterBuilder) Build(bus string) (*kernel.EmitterIO, error) {
	client, err := b.build(bus)
	if err != nil {
		return nil, err
	}
	return kernel.NewEmitter(client, client.config), nil
}

type ListenerBuilder struct{ *builder[*ListenerBuilder] }

func Listener() *ListenerBuilder {
	b := &ListenerBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *ListenerBuilder) Build() (*kernel.ListenerIO, error) {
	client, err := b.build("")
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
	service    EventBridge
}

// newBuilder creates new builder for EventBridge broker configuration.
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

// WithService configures AWS EventBridge client instance
func (b *builder[T]) WithService(service EventBridge) T {
	b.service = service
	return b.b
}

// build constructs the EventBridge client with configuration
func (b *builder[T]) build(bus string) (*Client, error) {
	client := &Client{
		config:  swarm.NewConfig(),
		bus:     bus,
		service: b.service,
	}

	if err := opts.Apply(&client.config, b.kernelOpts); err != nil {
		return nil, err
	}

	// Apply mandatory overrides
	client.config.PollFrequency = 5 * time.Microsecond

	if err := b.applyService(client); err != nil {
		return nil, err
	}

	if err := b.applyKernelSource(client); err != nil {
		return nil, err
	}

	return client, nil
}

func (b *builder[T]) applyService(c *Client) error {
	if c.service == nil {
		awsConfig, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return swarm.ErrServiceIO.With(err)
		}
		c.service = eventbridge.NewFromConfig(awsConfig)
	}
	return nil
}

func (b *builder[T]) applyKernelSource(c *Client) error {
	if c.config.Agent == "" {
		c.config.Agent = c.bus
	}
	return nil
}
