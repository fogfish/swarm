package queue

import (
	"encoding/json"
	"time"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Dequeue ...
*/
func Dequeue[T any](q swarm.Broker, queue ...string) (<-chan *swarm.Msg[T], chan<- *swarm.Msg[T]) {
	ch := swarm.NewMsgDeqCh[T]()

	cat := typeOf[T]()
	qid := cat
	if len(queue) > 0 {
		qid = queue[0]
	}

	sock, err := q.Dequeue(qid, ch)
	if err != nil {
		panic(err)
	}

	pipe.ForEach(ch.Ack, func(object *swarm.Msg[T]) {
		sock.Ack(swarm.Bag{
			Queue:    qid,
			Category: cat,
			Digest:   object.Digest,
		})
	})

	pipe.Emit(ch.Msg, 100*time.Millisecond, func() (*swarm.Msg[T], error) {
		bag, err := sock.Deq(cat)
		if err != nil {
			return nil, err
		}

		msg := &swarm.Msg[T]{Digest: bag.Digest}
		if err := json.Unmarshal(bag.Object, &msg.Object); err != nil {
			return nil, err
		}

		return msg, nil
	})

	return ch.Msg, ch.Ack
}
