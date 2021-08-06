package ephemeral

import (
	"context"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue"
)

/*

Ephemeral queue is in-memory unbound golang "channel"
See details about the pattern at
https://medium.com/capital-one-tech/building-an-unbounded-channel-in-go-789e175cd2cd

*/

type ephemeral struct {
	*queue.Queue
}

//
func New(sys swarm.System) (swarm.Queue, error) {
	q := &ephemeral{}

	// recv, send := q.create(sys)
	// q.Queue = queue.New(sys, recv, send)

	return q, nil
}

//
func (q *ephemeral) create(sys swarm.System) (<-chan *swarm.Message, chan<- *swarm.Message) {
	recv := make(chan *swarm.Message)
	send := make(chan *swarm.Message)

	sys.Spawn(func(ctx context.Context) {
		logger.Notice("start ephemeral queue %p", q)

		// TODO: linked list
		q := []*swarm.Message{}

		head := func() *swarm.Message {
			if len(q) == 0 {
				return nil
			}
			return q[0]
		}

		emit := func() chan<- *swarm.Message {
			if len(q) == 0 {
				return nil
			}
			return send
		}

		for {
			select {
			case <-ctx.Done():
				logger.Notice("stop ephemeral queue %p", q)
				return
			case v := <-recv:
				q = append(q, v)
			case emit() <- head():
				q = q[1:]
			}
		}
	})

	return send, recv
}
