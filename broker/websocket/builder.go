//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package websocket

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/fogfish/logger/x/xlog"
	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

const (
	// Environment variable defines event category (the route function is bound with)
	EnvConfigEventCategory = "CONFIG_SWARM_WS_EVENT_CATEGORY"

	// Environment variable defines URL to WebSocket API Gateway
	EnvConfigSourceWebSocketUrl = "CONFIG_SWARM_SOURCE_WS"
)

func Must[T any](v T, err error) T {
	if err != nil {
		xlog.Emergency("websocket broker has failed", err)
	}
	return v
}

type EndpointBuilder struct{ *builder[*EndpointBuilder] }

func Endpoint() *EndpointBuilder {
	b := &EndpointBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *EndpointBuilder) Build(endpoint string) (*kernel.Kernel, error) {
	client, err := b.build()
	if err != nil {
		return nil, err
	}

	if err := b.applyService(client, endpoint); err != nil {
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

func (b *EmitterBuilder) Build(endpoint string) (*kernel.EmitterCore, error) {
	client, err := b.build()
	if err != nil {
		return nil, err
	}

	if err := b.applyService(client, endpoint); err != nil {
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

func (b *ListenerBuilder) Build() (*kernel.ListenerCore, error) {
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
	service    Gateway
}

// newBuilder creates new builder for WebSocket broker configuration.
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

// WithService configures AWS API Gateway Management API client instance
func (b *builder[T]) WithService(service Gateway) T {
	b.service = service
	return b.b
}

// build constructs the WebSocket client with configuration
func (b *builder[T]) build() (*Client, error) {
	client := &Client{
		config:  swarm.NewConfig(),
		service: b.service,
	}

	if err := opts.Apply(&client.config, b.kernelOpts); err != nil {
		return nil, err
	}

	client.config.PollFrequency = 5 * time.Microsecond

	return client, nil
}

func (b *builder[T]) applyService(c *Client, endpoint string) error {
	if c.service == nil {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return swarm.ErrServiceIO.With(err)
		}

		c.service = apigatewaymanagementapi.NewFromConfig(cfg,
			func(o *apigatewaymanagementapi.Options) {
				if strings.HasPrefix(endpoint, "wss://") {
					endpoint = strings.Replace(endpoint, "wss://", "https://", 1)
				}

				if strings.HasPrefix(endpoint, "ws://") {
					endpoint = strings.Replace(endpoint, "ws://", "http://", 1)
				}

				o.BaseEndpoint = aws.String(endpoint)
			},
		)
	}
	return nil
}
