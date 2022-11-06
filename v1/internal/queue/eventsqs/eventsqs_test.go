//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/fogfish/swarm"
	sutRecv "github.com/fogfish/swarm/internal/queue/eventsqs"
	"github.com/fogfish/swarm/internal/queue/qtest"
)

func TestEventSQS(t *testing.T) {
	qtest.TestDequeue(t, mkDequeue)
}

func mkDequeue(
	sys swarm.System,
	policy *swarm.Policy,
	eff chan string,
	returnCategory string,
	returnMessage string,
	returnReceipt string,
) swarm.Dequeue {
	deq := sutRecv.NewDequeue(sys, "test-bridge", policy)
	deq.Mock(mockLambda(eff, returnCategory, returnMessage, returnReceipt))
	return deq
}

/*

mock AWS Lambda Handler

*/
func mockLambda(
	loopback chan string,
	returnCategory string,
	returnMessage string,
	returnReceipt string,
) func(interface{}) {
	return func(handler interface{}) {
		msg, _ := json.Marshal(
			events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId:     "abc-def",
						ReceiptHandle: returnReceipt,
						Body:          returnMessage,
						MessageAttributes: map[string]events.SQSMessageAttribute{
							"Category": {StringValue: aws.String(returnCategory)},
						},
					},
				},
			},
		)

		h := lambda.NewHandler(handler)
		_, err := h.Invoke(context.Background(), msg)
		if err != nil {
			panic(err)
		}

		loopback <- returnReceipt
	}
}
