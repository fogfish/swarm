//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package sqs_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/queue/qtest"
	sut "github.com/fogfish/swarm/internal/queue/sqs"
	"github.com/fogfish/swarm/internal/system"
)

func TestSQS(t *testing.T) {
	qtest.TestEnqueue(t, mkQueue)
	qtest.TestDequeue(t, mkQueue)
}

//
//
func mkQueue(sys swarm.System, policy *swarm.Policy, eff chan string) (swarm.Enqueue, swarm.Dequeue) {
	awscli, err := system.NewSession()
	if err != nil {
		panic(err)
	}

	s := sut.NewEnqueue(sys, "test-sqs", policy, awscli)
	s.Mock(&mockSQS{loopback: eff})

	r := sut.NewDequeue(sys, "test-sqs", policy, awscli)
	r.Mock(&mockSQS{loopback: eff})

	return s, r
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

	if !strings.HasPrefix(*cat.StringValue, qtest.Category) {
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
					"Category": {StringValue: aws.String(qtest.Category)},
				},
				Body:          aws.String(qtest.Message),
				ReceiptHandle: aws.String(qtest.Receipt),
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

// func BenchmarkSend(b *testing.B) {
// 	msg := swarm.Bytes(strings.Repeat("x", 256))
// 	sys := swarm.New("bench")

// 	n := 4
// 	q := make([]chan<- swarm.Msg, n)

// 	for i := 0; i < n; i++ {
// 		mq := swarm.Must(sut.New(sys, "swarm-test"))
// 		q[i], _ = mq.Send(subject)
// 	}

// 	for i := 0; i < n; i++ {
// 		b.Run(fmt.Sprintf("n-%d", i), func(b *testing.B) {
// 			for n := 0; n < b.N; n++ {
// 				for j := 0; j <= i; j++ {
// 					q[j] <- msg
// 				}
// 			}
// 		})
// 	}

// 	sys.Stop()
// }

// func BenchmarkRecv(b *testing.B) {
// 	sys := swarm.New("bench")

// 	n := 4
// 	q := make([]<-chan swarm.Msg, n)
// 	a := make([]chan<- swarm.Msg, n)

// 	for i := 0; i < n; i++ {
// 		mq := swarm.Must(
// 			sut.New(sys, "swarm-test",
// 				sut.PolicyPoll(1*time.Nanosecond),
// 			),
// 		)
// 		q[i], a[i] = mq.Recv(subject)
// 	}

// 	for i := 0; i < n; i++ {
// 		b.Run(fmt.Sprintf("n-%d", i), func(b *testing.B) {
// 			m := make([]swarm.Msg, n)
// 			for n := 0; n < b.N; n++ {

// 				for j := 0; j <= i; j++ {
// 					m[j] = <-q[j]
// 				}

// 				for j := 0; j <= i; j++ {
// 					a[j] <- m[j]
// 				}

// 			}
// 		})
// 	}

// 	sys.Stop()
// }
