package bytes

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Enqueue creates pair of channels to send messages and dead-letter queue
*/
func Enqueue(q swarm.Broker, cat string) (chan<- []byte, <-chan []byte) {
	conf := q.Config()
	ch := swarm.NewMsgEnqCh[[]byte](conf.EnqueueCapacity)

	sock := q.Enqueue(cat, ch)

	pipe.ForEach(ch.Msg, func(object []byte) {
		bag := swarm.Bag{Category: cat, Object: object}
		err := conf.Backoff.Retry(func() error { return sock.Enq(bag) })
		if err != nil {
			ch.Err <- object
		}
	})

	return ch.Msg, ch.Err
}
