package queue

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Dequeue ...
*/
func Dequeue(q swarm.Broker, cat string) (<-chan *swarm.Msg[[]byte], chan<- *swarm.Msg[[]byte]) {
	ch := swarm.NewMsgDeqCh[[]byte]()

	sock, err := q.Dequeue(cat, ch)
	if err != nil {
		panic(err)
	}

	pipe.ForEach(ch.Ack, func(object *swarm.Msg[[]byte]) {
		sock.Ack(swarm.Bag{
			Category: cat,
			Digest:   object.Digest,
		})
	})

	pipe.Emit(ch.Msg, q.Config().PollFrequency, func() (*swarm.Msg[[]byte], error) {
		bag, err := sock.Deq(cat)
		if err != nil {
			return nil, err
		}

		msg := &swarm.Msg[[]byte]{
			Object: bag.Object,
			Digest: bag.Digest,
		}

		return msg, nil
	})

	return ch.Msg, ch.Ack
}
