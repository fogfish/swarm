package bytes

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Dequeue ...
*/
func Dequeue(q swarm.Broker, cat string) (<-chan *swarm.Msg[[]byte], chan<- *swarm.Msg[[]byte]) {
	conf := q.Config()
	ch := swarm.NewMsgDeqCh[[]byte](conf.DequeueCapacity)

	sock := q.Dequeue(cat, ch)

	pipe.ForEach(ch.Ack, func(object *swarm.Msg[[]byte]) {
		err := conf.Backoff.Retry(func() error {
			return sock.Ack(swarm.Bag{
				Category: cat,
				Digest:   object.Digest,
			})
		})
		if err != nil && conf.StdErr != nil {
			conf.StdErr <- err
		}
	})

	pipe.Emit(ch.Msg, q.Config().PollFrequency, func() (*swarm.Msg[[]byte], error) {
		var bag swarm.Bag
		err := conf.Backoff.Retry(func() (err error) {
			bag, err = sock.Deq(cat)
			return
		})
		if err != nil {
			if conf.StdErr != nil {
				conf.StdErr <- err
			}
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