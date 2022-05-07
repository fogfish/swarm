package eventbridge

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
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

	Bus   eventbridgeiface.EventBridgeAPI
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
func (q *Queue) MockSend(mock eventbridgeiface.EventBridgeAPI) {
	q.Bus = mock
}

//
func (q *Queue) MockRecv(mock func(interface{})) {
	q.Start = mock
}

//
func (q *Queue) newSession(defSession []*session.Session) error {
	if len(defSession) != 0 {
		q.Bus = eventbridge.New(defSession[0])
		return nil
	}

	awscli, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return fmt.Errorf("Failed to create aws client %w", err)
	}

	q.Bus = eventbridge.New(awscli)
	return nil
}

func (q *Queue) ID() string {
	return q.id
}

/*

spawnSendIO create go routine for emiting messages
*/
func (q *Queue) Send() (chan *swarm.Bag, error) {
	return adapter.Send(q.adapter, q.send), nil
}

func (q *Queue) send(msg *swarm.Bag) error {
	ret, err := q.Bus.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{
			{
				EventBusName: aws.String(msg.System),
				Source:       aws.String(msg.Queue),
				DetailType:   aws.String(msg.Category),
				Detail:       aws.String(string(msg.Object.Bytes())),
			},
		},
	})

	if err != nil {
		return err
	}

	if *ret.FailedEntryCount > 0 {
		return fmt.Errorf("%v: %v", ret.Entries[0].ErrorCode, ret.Entries[0].ErrorMessage)
	}

	return nil
}

func (q *Queue) Recv() (chan *swarm.Bag, error) {
	sock := make(chan *swarm.Bag)
	conf := make(chan *swarm.Bag)

	go func() {
		logger.Notice("init aws eventbridge recv %s", q.id)

		// Note: this indirect synonym for lambda.Start
		q.Start(
			func(evt events.CloudWatchEvent) error {
				logger.Debug("cloudwatch event %+v", evt)

				sock <- &swarm.Bag{
					Category: evt.DetailType,
					Queue:    evt.Source,
					Object: &swarm.Msg{
						Payload: evt.Detail,
						Receipt: evt.ID,
					},
				}

				select {
				case <-conf:
					return nil
				case <-time.After(q.adapter.Policy.TimeToFlight):
					return errors.New("message ack timeout: " + evt.ID)
				}
			},
		)
		q.sys.Stop()
	}()

	q.qconf = conf
	return sock, nil
}

func (q *Queue) Conf() (chan *swarm.Bag, error) {
	return q.qconf, nil
}
