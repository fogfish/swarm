package queue

import (
	"encoding/json"

	"github.com/fogfish/swarm"
)

type Queue[T any] interface {
	Enqueue(T) error
}

//
type queue[T any] struct {
	cat  string
	conf *swarm.Config
	sock swarm.Enqueue
}

func (q queue[T]) Sync()  {}
func (q queue[T]) Close() {}

func (q queue[T]) Enqueue(object T) error {
	msg, err := json.Marshal(object)
	if err != nil {
		return err
	}

	bag := swarm.Bag{Category: q.cat, Object: msg}
	err = q.conf.Backoff.Retry(func() error { return q.sock.Enq(bag) })
	if err != nil {
		return err
	}

	return nil
}

//
func New[T any](q swarm.Broker, category ...string) Queue[T] {
	cat := typeOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	queue := &queue[T]{cat: cat, conf: q.Config()}
	sock, err := q.Enqueue(cat, queue)
	if err != nil {
		panic(err)
	}
	queue.sock = sock

	return queue
}
