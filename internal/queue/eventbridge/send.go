package eventbridge

import (
	"fmt"

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
type Sender struct {
	id      string
	adapter *adapter.Adapter
	sock    chan *swarm.Bag

	sys swarm.System

	client eventbridgeiface.EventBridgeAPI
}

/*

New ...
*/
func NewSender(
	sys swarm.System,
	queue string,
	policy *swarm.Policy,
	session *session.Session,
) *Sender {
	logger := logger.With(logger.Note{"type": "eventbridge", "q": queue})
	adapt := adapter.New(sys, policy, logger)

	return &Sender{
		id:      queue,
		adapter: adapt,

		sys: sys,

		client: eventbridge.New(session),
	}
}

//
func (q *Sender) Mock(mock eventbridgeiface.EventBridgeAPI) {
	q.client = mock
}

//
func (q *Sender) ID() string {
	return q.id
}

//
func (q *Sender) Start() error {
	return nil
}

//
func (q *Sender) Close() error {
	close(q.sock)

	return nil
}

/*

spawnSendIO create go routine for emiting messages
*/
func (q *Sender) Send() chan *swarm.Bag {
	if q.sock == nil {
		q.sock = adapter.Send(q.adapter, q.send)
	}
	return q.sock
}

func (q *Sender) send(msg *swarm.Bag) error {
	ret, err := q.client.PutEvents(&eventbridge.PutEventsInput{
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
