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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// SQS
type SQS interface {
	GetQueueUrl(context.Context, *sqs.GetQueueUrlInput, ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	SendMessage(context.Context, *sqs.SendMessageInput, ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type Client struct {
	service SQS
	config  swarm.Config
	queue   *string
	isFIFO  bool
}

func New(queue string, opts ...swarm.Option) (swarm.Broker, error) {
	cli, err := NewSQS(queue, opts...)
	if err != nil {
		return nil, err
	}

	config := swarm.NewConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return kernel.New(cli, cli, config), err
}

func NewSQS(queue string, opts ...swarm.Option) (*Client, error) {
	config := swarm.NewConfig()
	for _, opt := range opts {
		opt(&config)
	}

	api, err := newService(&config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.NetworkTimeout)
	defer cancel()

	spec, err := api.GetQueueUrl(ctx,
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(queue),
		},
	)
	if err != nil {
		return nil, swarm.ErrServiceIO.New(err)
	}

	return &Client{
		service: api,
		config:  config,
		queue:   spec.QueueUrl,
		isFIFO:  strings.HasSuffix(queue, ".fifo"),
	}, nil
}

func newService(conf *swarm.Config) (SQS, error) {
	if conf.Service != nil {
		service, ok := conf.Service.(SQS)
		if ok {
			return service, nil
		}
	}

	aws, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, swarm.ErrServiceIO.New(err)
	}

	return sqs.NewFromConfig(aws), nil
}

// Enq enqueues message to broker
func (cli *Client) Enq(bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout)
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
		return swarm.ErrEnqueue.New(err)
	}

	return nil
}

func (cli *Client) Ack(digest string) error {
	ctx, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout)
	defer cancel()

	_, err := cli.service.DeleteMessage(ctx,
		&sqs.DeleteMessageInput{
			QueueUrl:      cli.queue,
			ReceiptHandle: aws.String(digest),
		},
	)
	if err != nil {
		return swarm.ErrServiceIO.New(err)
	}

	return nil
}

// Deq dequeues message from broker
func (cli Client) Ask() ([]swarm.Bag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout*2)
	defer cancel()

	result, err := cli.service.ReceiveMessage(ctx,
		&sqs.ReceiveMessageInput{
			MessageAttributeNames: []string{string(types.QueueAttributeNameAll)},
			QueueUrl:              cli.queue,
			MaxNumberOfMessages:   1, // TODO
			WaitTimeSeconds:       int32(cli.config.NetworkTimeout.Seconds()),
		},
	)
	if err != nil {
		return nil, swarm.ErrDequeue.New(err)
	}

	if len(result.Messages) == 0 {
		return nil, nil
	}

	head := result.Messages[0]

	return []swarm.Bag{
		{
			Category: attr(&head, "Category"),
			Object:   []byte(*head.Body),
			Digest:   swarm.Digest{Brief: *head.ReceiptHandle},
		},
	}, nil
}

func attr(msg *types.Message, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}
