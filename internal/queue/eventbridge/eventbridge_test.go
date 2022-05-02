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
	"github.com/fogfish/swarm"
	sut "github.com/fogfish/swarm/internal/queue/eventbridge"
	"github.com/fogfish/swarm/internal/queue/qtest"
)

// const (
// 	subject = "eventbridge.test"
// 	message = "{\"some\":\"message\"}"
// )

// func TestNew(t *testing.T) {
// 	side := make(chan string)
// 	sys := swarm.New("test")

// 	swarm.Must(
// 		sut.New(sys, "swarm-test",
// 			mock(side),
// 			sut.PolicyIO(backoff.Const(1*time.Second, 3)),
// 		),
// 	)

// 	sys.Stop()
// }

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

func TestSend(t *testing.T) {
	qtest.TestSend(t, mkSend)
	qtest.TestRecv(t, mkRecv)

	// side := make(chan string)
	// sys := queue.System("test")
	// sutq, _ := sut.New(sys, "sutq", swarm.DefaultPolicy())
	// sutq.Mock(&mockEventBridge{loopback: side})

	// queue := sys.Queue(sutq)
	// out, _ := queue.Send(subject)

	// // t.Run("Success", func(t *testing.T) {
	// // 	out, _ := queue.Send(subject)
	// out <- swarm.Bytes(message)

	// go sys.Wait()
	// sys.Stop()

	// it.Ok(t).
	// 	If(<-side).Equal(message)
	// })

	// t.Run("Failure", func(t *testing.T) {
	// 	out, err := queue.Send("other")
	// 	out <- swarm.Bytes(message)

	// 	it.Ok(t).
	// 		If(<-err).Equal(swarm.Bytes(message))
	// })

	// queue.Wait()
	// sys.Stop()
}

// func TestRecv(t *testing.T) {
// 	side := make(chan string)

// 	sys := swarm.New("test")
// 	queue := swarm.Must(sut.New(sys, "test", mockLambda(side)))

// 	t.Run("Success", func(t *testing.T) {
// 		msg, ack := queue.Recv(subject)
// 		val := <-msg
// 		ack <- val

// 		it.Ok(t).
// 			If(string(val.Bytes())).Equal(message).
// 			If(<-side).Equal("ack")
// 	})

// 	sys.Stop()
// }

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
		time.Sleep(1 * time.Second)
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

		loopback <- "ack"
	}
}
