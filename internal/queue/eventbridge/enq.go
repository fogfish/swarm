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
type Enqueue struct {
	id      string
	adapter *adapter.Adapter
	sock    chan *swarm.BagStdErr

	sys swarm.System

	client eventbridgeiface.EventBridgeAPI
	logger logger.Logger
}

/*

New ...
*/
func NewEnqueue(
	sys swarm.System,
	queue string,
	policy *swarm.Policy,
	session *session.Session,
) *Enqueue {
	logger := logger.With(
		logger.Note{
			"type":  "eventbridge",
			"queue": sys.ID() + "://" + queue,
		},
	)

	adapt := adapter.New(policy, logger)
	return &Enqueue{
		id:      queue,
		adapter: adapt,

		sys: sys,

		client: eventbridge.New(session),
		logger: logger,
	}
}

//
func (q *Enqueue) Mock(mock eventbridgeiface.EventBridgeAPI) {
	q.client = mock
}

//
func (q *Enqueue) Listen() error {
	q.logger.Info("enqueue listening")

	return nil
}

//
func (q *Enqueue) Close() error {
	close(q.sock)

	q.logger.Info("enqueue closed")
	return nil
}

/*

spawnSendIO create go routine for emiting messages
*/
func (q *Enqueue) Enq() chan *swarm.BagStdErr {
	if q.sock == nil {
		q.sock = adapter.Enq(q.adapter, q.EnqSync)
	}
	return q.sock
}

func (q *Enqueue) EnqSync(msg *swarm.Bag) error {
	ret, err := q.client.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{
			{
				EventBusName: aws.String(msg.System),
				Source:       aws.String(msg.Queue),
				DetailType:   aws.String(msg.Category),
				Detail:       aws.String(string(msg.Object)),
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
