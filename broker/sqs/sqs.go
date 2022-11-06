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
	queue   *string
}

func newClient(service SQS, queue string) (*client, error) {
	api, err := newService(service)
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
		queue:   spec.QueueUrl,
	}, nil
}

func newService(service SQS) (SQS, error) {
	if service != nil {
		return service, nil
	}

	aws, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return sqs.NewFromConfig(aws), nil
}

// Enq enqueues message to broker
func (cli *client) Enq(bag swarm.Bag) error {
	_, err := cli.service.SendMessage(
		context.TODO(),
		&sqs.SendMessageInput{
			MessageAttributes: map[string]types.MessageAttributeValue{
				"Queue":    {StringValue: aws.String(bag.Queue), DataType: aws.String("String")},
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
	_, err := cli.service.DeleteMessage(
		context.TODO(),
		&sqs.DeleteMessageInput{
			QueueUrl:      cli.queue,
			ReceiptHandle: aws.String(string(bag.Digest)),
		},
	)
	return err
}

// Deq dequeues message from broker
func (cli client) Deq(cat string) (swarm.Bag, error) {
	result, err := cli.service.ReceiveMessage(
		context.TODO(),
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
		Queue:    attr(&head, "Queue"),
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
