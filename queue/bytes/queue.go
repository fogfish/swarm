package bytes

import (
	"github.com/fogfish/swarm"
)

type Queue interface {
	Enqueue([]byte) error
}

//
type queue struct {
	cat  string
	conf swarm.Config
	sock swarm.Enqueue
}

func (q queue) Sync()  {}
func (q queue) Close() {}

func (q queue) Enqueue(object []byte) error {
	bag := swarm.Bag{Category: q.cat, Object: object}
	err := q.conf.Backoff.Retry(func() error { return q.sock.Enq(bag) })
	if err != nil {
		return err
	}

	return nil
}

//
func New(q swarm.Broker, category string) Queue {
	queue := &queue{cat: category, conf: q.Config()}
	queue.sock = q.Enqueue(category, queue)

	return queue
}
