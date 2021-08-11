package sqs_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/fogfish/it"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/backoff"
	sut "github.com/fogfish/swarm/queue/sqs"
)

const (
	subject = "sqs.test"
	message = "some message"
)

func TestNew(t *testing.T) {
	side := make(chan string)
	sys := swarm.New("test")

	swarm.Must(
		sut.New(sys, "swarm-test",
			mock(side),
			sut.PolicyIO(backoff.Const(1*time.Second, 3)),
			sut.PolicyPoll(1*time.Second),
		),
	)

	sys.Stop()
}

func TestSend(t *testing.T) {
	side := make(chan string)
	sys := swarm.New("test")
	queue := swarm.Must(sut.New(sys, "test", mock(side)))

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
	queue := swarm.Must(sut.New(sys, "test", mock(side)))

	t.Run("Success", func(t *testing.T) {
		msg, _ := queue.Recv(subject)
		val := <-msg

		it.Ok(t).
			If(val.Bytes()).Equal([]byte(message))
	})

	t.Run("Acknowledge", func(t *testing.T) {
		msg, ack := queue.Recv(subject)
		val := <-msg
		ack <- val

		it.Ok(t).
			If(<-side).Equal("ack")
	})

	sys.Stop()
	time.Sleep(3 * time.Second)
}

/*

mock AWS SQS

*/
func mock(loopback chan string) sut.Config {
	return func(q *sut.Queue) {
		q.SQS = &mockSQS{
			loopback: loopback,
		}
	}
}

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

	if *cat.StringValue != subject {
		return nil, fmt.Errorf("Bad message category")
	}

	m.loopback <- aws.StringValue(s.MessageBody)
	return &sqs.SendMessageOutput{}, nil
}

func (m *mockSQS) ReceiveMessage(s *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	return &sqs.ReceiveMessageOutput{
		Messages: []*sqs.Message{
			{
				MessageAttributes: map[string]*sqs.MessageAttributeValue{
					"Category": {StringValue: aws.String(subject)},
				},
				Body:          aws.String(message),
				ReceiptHandle: aws.String("ack"),
			},
		},
	}, nil
}

func (m *mockSQS) DeleteMessage(s *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	m.loopback <- aws.StringValue(s.ReceiptHandle)
	return &sqs.DeleteMessageOutput{}, nil
}

/*

Benchmark: 256 bytes

*/

func BenchmarkSend(b *testing.B) {
	msg := swarm.Bytes(strings.Repeat("x", 256))
	sys := swarm.New("test")
	queue := swarm.Must(sut.New(sys, "test"))
	send, _ := queue.Send(subject)
	// recv, acks := queue.Recv(subject)

	b.Run("", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			send <- msg
			// acks <- <-recv
		}
	})

	sys.Stop()
}
