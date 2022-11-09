package queue

import (
	"encoding/json"
	"reflect"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Enqueue creates pair of channels to send messages and dead-letter queue
*/
func Enqueue[T any](q swarm.Broker, category ...string) (chan<- T, <-chan T) {
	ch := swarm.NewMsgEnqCh[T](q.Config().EnqueueCapacity)

	cat := typeOf[T]()
	if len(category) > 0 {
		cat = category[0]
	}

	sock, err := q.Enqueue(cat, ch)
	if err != nil {
		panic(err)
	}

	pipe.ForEach(ch.Msg, func(object T) {
		msg, err := json.Marshal(object)
		if err != nil {
			ch.Err <- object
			return
		}

		bag := swarm.Bag{Category: cat, Object: msg}

		if err := sock.Enq(bag); err != nil {
			ch.Err <- object
		}
	})

	return ch.Msg, ch.Err
}

func typeOf[T any]() string {
	//
	// TODO: fix
	//   Action[*swarm.User] if container type is used
	//

	typ := reflect.TypeOf(*new(T))
	cat := typ.Name()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().Name()
	}

	return cat
}

//
func Must(broker swarm.Broker, err error) swarm.Broker {
	if err != nil {
		panic(err)
	}

	return broker
}
