//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fogfish/swarm"
)

type client struct {
	service SQS
	config  *swarm.Config
	queue   *string
}

func newClient(queue string, config *swarm.Config) (*client, error) {
	api, err := newService(config)
	if err != nil {
		return nil, err
	}

	spec, err := api.GetQueueUrl(
		context.TODO(),
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(queue),
		},
	)
	if err != nil {
		return nil, err
	}

	return &client{
		service: api,
		config:  config,
		queue:   spec.QueueUrl,
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
		return nil, err
	}

	return sqs.NewFromConfig(aws), nil
}

// Enq enqueues message to broker
func (cli *client) Enq(bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout)
	defer cancel()

	_, err := cli.service.SendMessage(ctx,
		&sqs.SendMessageInput{
			MessageAttributes: map[string]types.MessageAttributeValue{
				"Source":   {StringValue: aws.String(cli.config.Source), DataType: aws.String("String")},
				"Category": {StringValue: aws.String(bag.Category), DataType: aws.String("String")},
			},
			MessageBody: aws.String(string(bag.Object)),
			QueueUrl:    cli.queue,
		},
	)
	return err
}

// Ack acknowledges message to broker
func (cli *client) Ack(bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout)
	defer cancel()

	_, err := cli.service.DeleteMessage(ctx,
		&sqs.DeleteMessageInput{
			QueueUrl:      cli.queue,
			ReceiptHandle: aws.String(string(bag.Digest)),
		},
	)
	return err
}

// Deq dequeues message from broker
func (cli client) Deq(cat string) (swarm.Bag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout)
	defer cancel()

	result, err := cli.service.ReceiveMessage(ctx,
		&sqs.ReceiveMessageInput{
			MessageAttributeNames: []string{string(types.QueueAttributeNameAll)},
			QueueUrl:              cli.queue,
			MaxNumberOfMessages:   1,  // TODO
			WaitTimeSeconds:       10, // TODO
		},
	)
	if err != nil {
		return swarm.Bag{}, err
	}

	if len(result.Messages) == 0 {
		return swarm.Bag{}, nil
	}

	head := result.Messages[0]

	return swarm.Bag{
		Category: attr(&head, "Category"),
		Object:   []byte(*head.Body),
		Digest:   *head.ReceiptHandle,
	}, nil
}

func attr(msg *types.Message, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}
