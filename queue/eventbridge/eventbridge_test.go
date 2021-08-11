package eventbridge_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"github.com/fogfish/it"
	"github.com/fogfish/swarm"
	sut "github.com/fogfish/swarm/queue/eventbridge"
)

const (
	subject = "eventbridge.test"
	message = "{\"some\": \"message\"}"
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

/*
func TestSend(t *testing.T) {

	sys := swarm.New("test")
	queue := swarm.Must(sut.New(sys, "swarm-test"))

	a, _ := queue.Send("eventbridge.test.a")
	b, _ := queue.Send("eventbridge.test.b")
	c, _ := queue.Send("eventbridge.test.c")

	a <- swarm.Bytes(message)
	b <- swarm.Bytes(message)
	c <- swarm.Bytes(message)

	time.Sleep(2 * time.Second)
	sys.Stop()
	time.Sleep(3 * time.Second)
}
*/

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
