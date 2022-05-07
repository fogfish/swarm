package eventsqs

import (
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/queue/adapter"
)

/*

Queue ...
*/
type Queue struct {
	adapter *adapter.Adapter
	id      string
	sys     swarm.System

	qrecv chan *swarm.Bag
	qconf chan *swarm.Bag

	url   *string
	SQS   sqsiface.SQSAPI
	Start func(interface{})
}

/*

New ...
*/
func New(
	sys swarm.System,
	id string,
	policy *swarm.Policy,
	defSession ...*session.Session,
) (*Queue, error) {
	logger := logger.With(logger.Note{
		"type": "eventbridge",
		"q":    id,
	})
	q := &Queue{
		id:      id,
		sys:     sys,
		adapter: adapter.New(sys, policy, logger),
		Start:   lambda.Start,
	}
	if err := q.newSession(defSession); err != nil {
		return nil, err
	}

	return q, nil
}

//
func (q *Queue) Mock(mock sqsiface.SQSAPI) {
	q.SQS = mock
}

//
func (q *Queue) MockLambda(mock func(interface{})) {
	q.Start = mock
}

//
func (q *Queue) newSession(defSession []*session.Session) error {
	if len(defSession) != 0 {
		q.SQS = sqs.New(defSession[0])
		return nil
	}

	awscli, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return fmt.Errorf("Failed to create aws client %w", err)
	}

	q.SQS = sqs.New(awscli)
	return nil
}

func (q *Queue) ID() string {
	return q.id
}

//
func (q *Queue) lookupQueue(id string) error {
	spec, err := q.SQS.GetQueueUrl(
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(id),
		},
	)

	if err != nil {
		return fmt.Errorf("Failed to discover %s aws sqs: %w", id, err)
	}

	q.url = spec.QueueUrl
	return nil
}

/*

spawnSendIO create go routine for emiting messages
*/
func (q *Queue) Send() (chan *swarm.Bag, error) {
	if q.url == nil {
		if err := q.lookupQueue(q.id); err != nil {
			return nil, err
		}
	}

	return adapter.Send(q.adapter, q.send), nil
}

func (q *Queue) send(msg *swarm.Bag) error {
	_, err := q.SQS.SendMessage(
		&sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"Category": {StringValue: aws.String(msg.Category), DataType: aws.String("String")},
				"System":   {StringValue: aws.String(msg.System), DataType: aws.String("String")},
				"Queue":    {StringValue: aws.String(msg.Queue), DataType: aws.String("String")},
			},
			MessageBody: aws.String(string(msg.Object.Bytes())),
			QueueUrl:    q.url,
		},
	)

	return err
}

func (q *Queue) Recv() (chan *swarm.Bag, error) {
	if q.url == nil {
		if err := q.lookupQueue(q.id); err != nil {
			return nil, err
		}
	}

	sock := make(chan *swarm.Bag)
	conf := make(chan *swarm.Bag)

	go func() {
		logger.Notice("init aws eventbridge recv %s", q.id)

		// Note: this indirect synonym for lambda.Start
		q.Start(
			func(events events.SQSEvent) error {
				logger.Debug("cloudwatch event %+v", events)

				acks := map[string]bool{}

				for _, evt := range events.Records {
					acks[evt.ReceiptHandle] = false
					sock <- &swarm.Bag{
						Category: attr(&evt, "Category"),
						System:   attr(&evt, "System"),
						Queue:    attr(&evt, "Queue"),
						Object: &swarm.Msg{
							Payload: []byte(evt.Body),
							Receipt: evt.ReceiptHandle,
						},
					}
				}

				for {
					select {
					case bag := <-conf:
						switch msg := bag.Object.(type) {
						case *swarm.Msg:
							delete(acks, msg.Receipt)
							if len(acks) == 0 {
								return nil
							}
						default:
							return fmt.Errorf("unsupported conf type %v", bag)
						}
					case <-time.After(q.adapter.Policy.TimeToFlight):
						return fmt.Errorf("sqs message ack timeout")
					}
				}

			},
		)
		q.sys.Stop()
	}()

	q.qconf = conf
	return sock, nil
}

func attr(msg *events.SQSMessage, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}

func (q *Queue) Conf() (chan *swarm.Bag, error) {
	return q.qconf, nil
}
