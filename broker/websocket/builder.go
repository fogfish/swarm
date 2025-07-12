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

// Environment variable to config event source
const EnvConfigEventType = "CONFIG_SWARM_WS_EVENT_TYPE"
const EnvConfigSourceWebSocket = "CONFIG_SWARM_WS_URL"

// Builder provides API for configuring WebSocket broker
type Builder struct {
	kernelOpts []opts.Option[swarm.Config]
	service    Gateway
}

// Channels creates new builder for WebSocket broker configuration.
func Channels() *Builder {
	kopts := []opts.Option[swarm.Config]{
		swarm.WithLogStdErr(),
		swarm.WithConfigFromEnv(),
	}
	// if val := os.Getenv(EnvConfigSourceWebSocket); val != "" {
	// 	kopts = append(kopts, swarm.WithSource(val))
	// }

	return &Builder{
		kernelOpts: kopts,
	}
}

// WithKernel configures swarm kernel options for advanced usage.
func (b *Builder) WithKernel(opts ...opts.Option[swarm.Config]) *Builder {
	b.kernelOpts = append(b.kernelOpts, opts...)
	return b
}

// WithService configures AWS API Gateway Management API client instance
func (b *Builder) WithService(service Gateway) *Builder {
	b.service = service
	return b
}

// NewEnqueuer creates enqueue routine to AWS API Gateway WebSocket
func (b *Builder) NewEnqueuer(endpoint string) (*kernel.EmitterCore, error) {
	client, err := b.build()
	if err != nil {
		return nil, err
	}

	if err := b.applyService(client, endpoint); err != nil {
		return nil, err
	}

	return kernel.NewEmitter(client, client.config), nil
}

// Creates enqueue routine to AWS API Gateway WebSocket
func (b *Builder) MustEnqueuer(endpoint string) *kernel.EmitterCore {
	client, err := b.NewEnqueuer(endpoint)
	if err != nil {
		xlog.Emergency("websocket client has failed", err)
		return nil
	}
	return client
}

// NewDequeuer creates dequeue routine from AWS API Gateway WebSocket (Lambda)
func (b *Builder) NewDequeuer() (*kernel.ListenerCore, error) {
	client, err := b.build()
	if err != nil {
		return nil, err
	}

	bridge := &bridge{kernel.NewBridge(client.config.TimeToFlight)}
	return kernel.NewListener(bridge, client.config), nil
}

// Creates dequeue routine from AWS API Gateway WebSocket (Lambda)
func (b *Builder) MustDequeuer() *kernel.ListenerCore {
	client, err := b.NewDequeuer()
	if err != nil {
		xlog.Emergency("websocket client has failed", err)
		return nil
	}
	return client
}

// NewClient creates duplex client for AWS API Gateway WebSocket (enqueue & dequeue)
func (b *Builder) NewClient(endpoint string) (*kernel.Kernel, error) {
	client, err := b.build()
	if err != nil {
		return nil, err
	}

	if err := b.applyService(client, endpoint); err != nil {
		return nil, err
	}

	bridge := &bridge{kernel.NewBridge(client.config.TimeToFlight)}

	return kernel.New(
		kernel.NewEmitter(client, client.config),
		kernel.NewListener(bridge, client.config),
	), nil
}

// Creates duplex client for AWS API Gateway WebSocket (enqueue & dequeue)
func (b *Builder) MustClient(endpoint string) *kernel.Kernel {
	client, err := b.NewClient(endpoint)
	if err != nil {
		xlog.Emergency("websocket client has failed", err)
		return nil
	}
	return client
}

func (b *Builder) build() (*Client, error) {
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

func (b *Builder) applyService(c *Client, endpoint string) error {
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
