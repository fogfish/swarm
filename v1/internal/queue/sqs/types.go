//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

//
// The file declares common types used by SQS adapter
//

package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type GetQueueUrl interface {
	GetQueueUrl(context.Context, *sqs.GetQueueUrlInput, ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
}

type SendMessage interface {
	SendMessage(context.Context, *sqs.SendMessageInput, ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type ReceiveMessage interface {
	ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
}

type DeleteMessage interface {
	DeleteMessage(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type EnqueueService interface {
	GetQueueUrl
	SendMessage
}

type DequeueService interface {
	GetQueueUrl
	ReceiveMessage
	DeleteMessage
}
