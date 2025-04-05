//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package sqs

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fogfish/opts"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// SQS
type SQS interface {
	GetQueueUrl(context.Context, *sqs.GetQueueUrlInput, ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	SendMessage(context.Context, *sqs.SendMessageInput, ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type Client struct {
	service   SQS
	config    swarm.Config
	queue     *string
	isFIFO    bool
	batchSize int
}

// Create enqueue routine to AWS SQS
func NewEnqueuer(queue string, opts ...Option) (*kernel.Enqueuer, error) {
	cli, err := newSQS(queue, opts...)
	if err != nil {
		return nil, err
	}

	return kernel.NewEnqueuer(cli, cli.config), nil
}

// Create dequeue routine to AWS SQS
func NewDequeuer(queue string, opts ...Option) (*kernel.Dequeuer, error) {
	cli, err := newSQS(queue, opts...)
	if err != nil {
		return nil, err
	}

	return kernel.NewDequeuer(cli, cli.config), nil
}

// Create enqueue & dequeue routine to AWS SQS
func New(queue string, opts ...Option) (*kernel.Kernel, error) {
	cli, err := newSQS(queue, opts...)
	if err != nil {
		return nil, err
	}

	return kernel.New(
		kernel.NewEnqueuer(cli, cli.config),
		kernel.NewDequeuer(cli, cli.config),
	), nil
}

func newSQS(queue string, opt ...Option) (*Client, error) {
	c := &Client{}
	if err := opts.Apply(c, defs); err != nil {
		return nil, err
	}
	if err := opts.Apply(c, opt); err != nil {
		return nil, err
	}

	if c.batchSize == 0 {
		c.batchSize = 1
	}

	if c.service == nil {
		aws, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, swarm.ErrServiceIO.With(err)
		}
		c.service = sqs.NewFromConfig(aws)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.NetworkTimeout)
	defer cancel()

	spec, err := c.service.GetQueueUrl(ctx,
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(queue),
		},
	)
	if err != nil {
		return nil, swarm.ErrServiceIO.With(err)
	}

	c.queue = spec.QueueUrl
	c.isFIFO = strings.HasSuffix(queue, ".fifo")

	return c, nil
}

// Enq enqueues message to broker
func (cli *Client) Enq(ctx context.Context, bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(ctx, cli.config.NetworkTimeout)
	defer cancel()

	var idMsgGroup *string
	if cli.isFIFO {
		idMsgGroup = aws.String(bag.Category)
	}

	_, err := cli.service.SendMessage(ctx,
		&sqs.SendMessageInput{
			MessageAttributes: map[string]types.MessageAttributeValue{
				"Source":   {StringValue: aws.String(cli.config.Source), DataType: aws.String("String")},
				"Category": {StringValue: aws.String(bag.Category), DataType: aws.String("String")},
			},
			MessageGroupId: idMsgGroup,
			MessageBody:    aws.String(string(bag.Object)),
			QueueUrl:       cli.queue,
		},
	)
	if err != nil {
		return swarm.ErrEnqueue.With(err)
	}

	return nil
}

func (cli *Client) Ack(ctx context.Context, digest string) error {
	ctx, cancel := context.WithTimeout(ctx, cli.config.NetworkTimeout)
	defer cancel()

	_, err := cli.service.DeleteMessage(ctx,
		&sqs.DeleteMessageInput{
			QueueUrl:      cli.queue,
			ReceiptHandle: aws.String(digest),
		},
	)
	if err != nil {
		return swarm.ErrServiceIO.With(err)
	}

	return nil
}

func (cli *Client) Err(ctx context.Context, digest string, err error) error {
	// Note: do nothing, AWS SQS makes the magic
	return nil
}

// Deq dequeues message from broker
func (cli Client) Ask(ctx context.Context) ([]swarm.Bag, error) {
	ctx, cancel := context.WithTimeout(ctx, cli.config.NetworkTimeout*2)
	defer cancel()

	result, err := cli.service.ReceiveMessage(ctx,
		&sqs.ReceiveMessageInput{
			MessageAttributeNames: []string{string(types.QueueAttributeNameAll)},
			QueueUrl:              cli.queue,
			MaxNumberOfMessages:   int32(cli.batchSize),
			WaitTimeSeconds:       int32(cli.config.NetworkTimeout.Seconds()),
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, nil
		}
		return nil, swarm.ErrDequeue.With(err)
	}

	if len(result.Messages) == 0 {
		return nil, nil
	}

	bag := make([]swarm.Bag, len(result.Messages))
	for i, msg := range result.Messages {
		bag[i] = swarm.Bag{
			Category: attr(&msg, "Category"),
			Digest:   aws.ToString(msg.ReceiptHandle),
			Object:   []byte(aws.ToString(msg.Body)),
		}
	}

	return bag, nil
}

func attr(msg *types.Message, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}
