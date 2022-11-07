package queue

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Enqueue creates pair of channels to send messages and dead-letter queue
*/
func Enqueue(q swarm.Broker, cat string) (chan<- []byte, <-chan []byte) {
	ch := swarm.NewMsgEnqCh[[]byte]()

	sock, err := q.Enqueue(cat, ch)
	if err != nil {
		panic(err)
	}

	pipe.ForEach(ch.Msg, func(object []byte) {
		err := sock.Enq(swarm.Bag{
			Category: cat,
			Object:   object,
		})
		if err != nil {
			ch.Err <- object
		}
	})

	return ch.Msg, ch.Err
}
