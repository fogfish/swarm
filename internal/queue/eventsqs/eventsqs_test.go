package eventsqs_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/fogfish/swarm"
	sut "github.com/fogfish/swarm/internal/queue/eventsqs"
	"github.com/fogfish/swarm/internal/queue/qtest"
)

func mkSend(sys swarm.System, policy *swarm.Policy, eff chan string) swarm.EventBus {
	q, err := sut.New(sys, "test-bridge", policy)
	if err != nil {
		panic(err)
	}
	q.Mock(&mockSQS{loopback: eff})
	return q
}

func mkRecv(sys swarm.System, policy *swarm.Policy, eff chan string) swarm.EventBus {
	q, err := sut.New(sys, "test-bridge", policy)
	if err != nil {
		panic(err)
	}
	q.Mock(&mockSQS{loopback: eff})
	q.MockLambda(mockLambda(eff))
	return q
}

func TestEventBridge(t *testing.T) {
	qtest.TestSend(t, mkSend)
	qtest.TestRecv(t, mkRecv)
}

//
//
type mockSQS struct {
	sqsiface.SQSAPI

	loopback chan string
}

func (m *mockSQS) GetQueueUrl(s *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/000000000000/mock"),
	}, nil
}

func (m *mockSQS) SendMessage(s *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	cat, exists := s.MessageAttributes["Category"]
	if !exists {
		return nil, fmt.Errorf("Bad message attributes")
	}

	if *cat.StringValue != qtest.Category {
		return nil, fmt.Errorf("Bad message category")
	}

	m.loopback <- aws.StringValue(s.MessageBody)
	return &sqs.SendMessageOutput{}, nil
}

/*

mock AWS Lambda Handler

*/
func mockLambda(loopback chan string) func(interface{}) {
	return func(handler interface{}) {
		msg, _ := json.Marshal(
			events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId:     "abc-def",
						ReceiptHandle: qtest.Receipt,
						Body:          qtest.Message,
						MessageAttributes: map[string]events.SQSMessageAttribute{
							"Category": {StringValue: aws.String(qtest.Category)},
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

		loopback <- qtest.Receipt
	}
}
