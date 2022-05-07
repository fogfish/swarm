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
	"github.com/fogfish/swarm/internal/system"
)

func mkQueue(sys swarm.System, policy *swarm.Policy, eff chan string) (swarm.Sender, swarm.Recver) {
	awscli, err := system.NewSession()
	if err != nil {
		panic(err)
	}

	s := sut.NewSender(sys, "test-bridge", policy, awscli)
	s.Mock(&mockEventBridge{loopback: eff})

	r := sut.NewRecver(sys, "test-bridge", policy)
	r.Mock(mockLambda(eff))
	return s, r
}

func TestEventBridge(t *testing.T) {
	qtest.TestSend(t, mkQueue)
	qtest.TestRecv(t, mkQueue)
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
