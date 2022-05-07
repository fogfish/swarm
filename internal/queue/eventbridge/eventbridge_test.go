package eventbridge_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"github.com/fogfish/swarm"
	sut "github.com/fogfish/swarm/internal/queue/eventbridge"
	"github.com/fogfish/swarm/internal/queue/qtest"
)

func mkSend(sys swarm.System, policy *swarm.Policy, eff chan string) swarm.EventBus {
	q, err := sut.New(sys, "test-bridge", policy)
	if err != nil {
		panic(err)
	}
	q.MockSend(&mockEventBridge{loopback: eff})
	return q
}

func mkRecv(sys swarm.System, policy *swarm.Policy, eff chan string) swarm.EventBus {
	q, err := sut.New(sys, "test-bridge", policy)
	if err != nil {
		panic(err)
	}
	q.MockRecv(mockLambda(eff))
	return q
}

func TestEventBridge(t *testing.T) {
	qtest.TestSend(t, mkSend)
	qtest.TestRecv(t, mkRecv)
}

type mockEventBridge struct {
	eventbridgeiface.EventBridgeAPI

	loopback chan string
}

func (m *mockEventBridge) PutEvents(s *eventbridge.PutEventsInput) (*eventbridge.PutEventsOutput, error) {
	if len(s.Entries) != 1 {
		return nil, fmt.Errorf("Bad request")
	}

	if aws.StringValue(s.Entries[0].DetailType) != qtest.Category {
		return nil, fmt.Errorf("Bad message category")
	}

	m.loopback <- aws.StringValue(s.Entries[0].Detail)
	return &eventbridge.PutEventsOutput{
		FailedEntryCount: aws.Int64(0),
	}, nil
}

/*

mock AWS Lambda Handler

*/
func mockLambda(loopback chan string) func(interface{}) {
	return func(handler interface{}) {
		msg, _ := json.Marshal(
			events.CloudWatchEvent{
				ID:         "abc-def",
				DetailType: qtest.Category,
				Detail:     json.RawMessage(qtest.Message),
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
