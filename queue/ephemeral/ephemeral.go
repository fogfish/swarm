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

//
func (q *ephemeral) create(sys swarm.System) (<-chan *queue.Bag, chan<- *queue.Bag) {
	recv := make(chan *queue.Bag)
	send := make(chan *queue.Bag)

	sys.Go(func(ctx context.Context) {
		logger.Notice("start ephemeral queue %s", q.ID)
		defer close(recv)
		defer close(send)

		// TODO: linked list
		mq := []*queue.Bag{}

		head := func() *queue.Bag {
			if len(mq) == 0 {
				return nil
			}
			return mq[0]
		}

		emit := func() chan<- *queue.Bag {
			if len(mq) == 0 {
				return nil
			}
			return send
		}

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
			case emit() <- head():
				mq = mq[1:]
			}
		}
	})

	return send, recv
}
