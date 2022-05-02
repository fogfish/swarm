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
func New(sys swarm.System, id string) (swarm.Queue, error) {
	q := &ephemeral{}

	recv, send := q.create(sys)

	q.Queue = queue.New(
		sys,
		id,
		func() (<-chan *queue.Bag, chan<- *queue.Bag) { return recv, nil },
		func() chan<- *queue.Bag { return send },
	)

	return q, nil
}

// TODO: linked list
type unbound []*queue.Bag

func (q unbound) head() *queue.Bag {
	if len(q) == 0 {
		return nil
	}
	return q[0]
}

func (q unbound) emit(ch chan<- *queue.Bag) chan<- *queue.Bag {
	if len(q) == 0 {
		return nil
	}
	return ch
}

// TODO: define collection of channel types
func (q *ephemeral) create(sys swarm.System) (<-chan *queue.Bag, chan<- *queue.Bag) {
	recv := make(chan *queue.Bag)
	send := make(chan *queue.Bag)

	sys.Go(func(ctx context.Context) {
		logger.Notice("start ephemeral queue %s", q.ID)
		defer close(recv)
		defer close(send)

		mq := unbound{}
		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("stop ephemeral queue %s", q.ID)
				return

			//
			case v := <-recv:
				mq = append(mq, v)

			//
			case mq.emit(send) <- mq.head():
				mq = mq[1:]
			}
		}
	})

	return send, recv
}
