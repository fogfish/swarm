package ephemeral

import (
	"context"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

/*

Ephemeral queue is in-memory unbound golang "channel"
See details about the pattern at
https://medium.com/capital-one-tech/building-an-unbounded-channel-in-go-789e175cd2cd

*/

type Queue struct {
	id   string
	send chan<- *swarm.Bag
	recv <-chan *swarm.Bag
	conf chan<- *swarm.Bag
}

//
func New(sys swarm.System, id string) (swarm.EventBus, error) {
	recv, send := create(sys, id)
	conf := blackhole(sys)

	return &Queue{
		id:   id,
		send: send,
		recv: recv,
		conf: conf,
	}, nil
}

// TODO: define collection of channel types
func create(sys swarm.System, id string) (<-chan *swarm.Bag, chan<- *swarm.Bag) {
	recv := make(chan *swarm.Bag)
	send := make(chan *swarm.Bag)

	sys.Go(func(ctx context.Context) {
		logger.Notice("start ephemeral queue %s", id)
		defer close(recv)
		defer close(send)

		mq := unbound{}
		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("stop ephemeral queue %s", id)
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

func blackhole(sys swarm.System) chan<- *swarm.Bag {
	conf := make(chan *swarm.Bag)

	sys.Go(func(ctx context.Context) {
		for {
			select {
			//
			case <-ctx.Done():
				return

			//
			case <-conf:
				continue
			}
		}
	})

	return conf
}

func (q *Queue) ID() string                       { return q.id }
func (q *Queue) Send() (chan<- *swarm.Bag, error) { return q.send, nil }
func (q *Queue) Recv() (<-chan *swarm.Bag, error) { return q.recv, nil }
func (q *Queue) Conf() (chan<- *swarm.Bag, error) { return q.conf, nil }

//
// TODO: linked list
type unbound []*swarm.Bag

func (q unbound) head() *swarm.Bag {
	if len(q) == 0 {
		return nil
	}
	return q[0]
}

func (q unbound) emit(ch chan<- *swarm.Bag) chan<- *swarm.Bag {
	if len(q) == 0 {
		return nil
	}
	return ch
}
