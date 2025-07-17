//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package sqs

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/fogfish/logger/x/xlog"
	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// Environment define name of source SQS queue
const EnvConfigSourceSQS = "CONFIG_SWARM_SOURCE_SQS"

// Environment define name of target SQS queue
const EnvConfigTargetSQS = "CONFIG_SWARM_TARGET_SQS"

func Must[T any](v T, err error) T {
	if err != nil {
		xlog.Emergency("sqs broker has failed", err)
	}
	return v
}

type EndpointBuilder struct{ *builder[*EndpointBuilder] }

func Endpoint() *EndpointBuilder {
	b := &EndpointBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *EndpointBuilder) Build(queue string) (*kernel.Kernel, error) {
	client, err := b.build(queue)
	if err != nil {
		return nil, err
	}
	return kernel.New(
		kernel.NewEmitter(client, client.config),
		kernel.NewListener(client, client.config),
	), nil
}

type EmitterBuilder struct{ *builder[*EmitterBuilder] }

func Emitter() *EmitterBuilder {
	b := &EmitterBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *EmitterBuilder) Build(queue string) (*kernel.EmitterIO, error) {
	client, err := b.build(queue)
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

func (b *ListenerBuilder) Build(queue string) (*kernel.ListenerIO, error) {
	client, err := b.build(queue)
	if err != nil {
		return nil, err
	}
	return kernel.NewListener(client, client.config), nil
}

//------------------------------------------------------------------------------

type builder[T any] struct {
	b           T
	kernelOpts  []opts.Option[swarm.Config]
	service     SQS
	batchSize   int
	askWaitTime time.Duration
}

// Channels creates new builder for SQS broker configuration.
func newBuilder[T any](b T) *builder[T] {
	kopts := []opts.Option[swarm.Config]{
		swarm.WithLogStdErr(),
		swarm.WithConfigFromEnv(),
	}

	return &builder[T]{
		b:           b,
		kernelOpts:  kopts,
		batchSize:   1,
		askWaitTime: 5 * time.Second,
	}
}

// WithKernel configures swarm kernel options for advanced usage.
func (b *builder[T]) WithKernel(opts ...opts.Option[swarm.Config]) T {
	b.kernelOpts = append(b.kernelOpts, opts...)
	return b.b
}

// WithService configures AWS SQS client instance
func (b *builder[T]) WithService(service SQS) T {
	b.service = service
	return b.b
}

// WithBatchSize configures SQS batch size.
// Note: AWS SQS limits the MaxNumberOfMessages to 10 per poller.
// This config increases number of poller for batch sizes > 10.
func (b *builder[T]) WithBatchSize(size int) T {
	b.batchSize = size
	return b.b
}

// WithWaitTime configures SQS message long polling.
func (b *builder[T]) WithWaitTime(duration time.Duration) T {
	b.askWaitTime = duration
	return b.b
}

// build constructs the SQS client with configuration
func (b *builder[T]) build(queue string) (*Client, error) { // Start with sensible defaults
	client := &Client{
		config:      swarm.NewConfig(),
		service:     b.service,
		batchSize:   b.batchSize,
		askWaitTime: b.askWaitTime,
	}

	if err := opts.Apply(&client.config, b.kernelOpts); err != nil {
		return nil, err
	}

	if err := b.applyBatchSize(client); err != nil {
		return nil, err
	}

	if err := b.applyService(client); err != nil {
		return nil, err
	}

	if err := b.applyQueue(client, queue); err != nil {
		return nil, err
	}

	return client, nil
}

func (b *builder[T]) applyBatchSize(c *Client) error {
	switch {
	case c.batchSize <= 0:
		c.batchSize = 1
	case c.batchSize > 10:
		const maxNumberOfMessages = 10
		c.config.PollerPool = b.batchSize/maxNumberOfMessages + 1
		c.batchSize = maxNumberOfMessages
	}

	return nil
}

func (b *builder[T]) applyService(c *Client) error {
	if c.service == nil {
		awsConfig, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return swarm.ErrServiceIO.With(err)
		}
		c.service = sqs.NewFromConfig(awsConfig)
	}
	return nil
}

func (b *builder[T]) applyQueue(c *Client, queue string) error {
	// Get queue URL
	ctx, cancel := context.WithTimeout(context.Background(), c.config.NetworkTimeout)
	defer cancel()

	spec, err := c.service.GetQueueUrl(ctx,
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(queue),
		},
	)
	if err != nil {
		return swarm.ErrServiceIO.With(err)
	}

	c.queue = spec.QueueUrl
	c.isFIFO = strings.HasSuffix(queue, ".fifo")

	return nil
}
