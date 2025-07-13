//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package embedded

import (
	"context"

	"github.com/fogfish/golem/pipe"
	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

type builder[T any] struct {
	b          T
	kernelOpts []opts.Option[swarm.Config]
}

type EndpointBuilder struct{ *builder[*EndpointBuilder] }

func Endpoint() *EndpointBuilder {
	b := &EndpointBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *EndpointBuilder) Build() (*kernel.Kernel, error) {
	client, err := b.build()
	if err != nil {
		return nil, err
	}

	return kernel.New(
		kernel.NewEmitter(client, client.config),
		kernel.NewListener(client, client.config),
	), nil
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

// build constructs the EventBridge client with configuration
func (b *builder[T]) build() (*Client, error) {
	client := &Client{
		config:  swarm.NewConfig(),
		context: context.Background(),
		bags:    make(map[string]*swarm.Bag),
	}

	if err := opts.Apply(&client.config, b.kernelOpts); err != nil {
		return nil, err
	}

	if err := b.applyService(client); err != nil {
		return nil, err
	}

	return client, nil
}

func (b *builder[T]) applyService(c *Client) error {
	cap := max(c.config.CapOut, c.config.CapRcv)
	c.recv, c.emit = pipe.New[*swarm.Bag](c.context, cap)
	return nil
}
