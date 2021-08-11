package eventbridge_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"github.com/fogfish/it"
	"github.com/fogfish/swarm"
	sut "github.com/fogfish/swarm/queue/eventbridge"
)

const (
	subject = "eventbridge.test"
	message = "{\"some\":\"message\"}"
)

func TestSend(t *testing.T) {
	side := make(chan string)
	sys := swarm.New("test")
	queue := swarm.Must(sut.New(sys, "swarm-test", mock(side)))

	t.Run("Success", func(t *testing.T) {
		out, _ := queue.Send(subject)
		out <- swarm.Bytes(message)

		it.Ok(t).
			If(<-side).Equal(message)
	})

	t.Run("Failure", func(t *testing.T) {
		out, err := queue.Send("other")
		out <- swarm.Bytes(message)

		it.Ok(t).
			If(<-err).Equal(swarm.Bytes(message))
	})

	sys.Stop()
	time.Sleep(3 * time.Second)
}

func TestRecv(t *testing.T) {
	side := make(chan string)

	sys := swarm.New("test")
	queue := swarm.Must(sut.New(sys, "test", mockLambda(side)))

	t.Run("Success", func(t *testing.T) {
		msg, ack := queue.Recv(subject)
		val := <-msg
		ack <- val

		it.Ok(t).
			If(string(val.Bytes())).Equal(message).
			If(<-side).Equal("ack")
	})

	sys.Stop()
}

/*

mock AWS SQS

*/
func mock(loopback chan string) sut.Config {
	return func(q *sut.Queue) {
		q.Bus = &mockEventBridge{
			loopback: loopback,
		}
	}
}

type mockEventBridge struct {
	eventbridgeiface.EventBridgeAPI

	loopback chan string
}

func (m *mockEventBridge) PutEvents(s *eventbridge.PutEventsInput) (*eventbridge.PutEventsOutput, error) {
	if len(s.Entries) != 1 {
		return nil, fmt.Errorf("Bad request")
	}

	if aws.StringValue(s.Entries[0].DetailType) != subject {
		return nil, fmt.Errorf("Bad message category")
	}

	m.loopback <- aws.StringValue(s.Entries[0].Detail)
	return &eventbridge.PutEventsOutput{}, nil
}

/*

mock AWS Lambda Handler

*/

func mockLambda(loopback chan string) sut.Config {
	return func(q *sut.Queue) {
		q.Start = func(handler interface{}) {
			time.Sleep(1 * time.Second)
			msg, _ := json.Marshal(
				events.CloudWatchEvent{
					ID:         "abc-def",
					DetailType: subject,
					Detail:     json.RawMessage(message),
				},
			)

			h := lambda.NewHandler(handler)
			_, err := h.Invoke(context.Background(), msg)
			if err != nil {
				panic(err)
			}

			loopback <- "ack"
		}
	}
}
