package queue

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fogfish/golem/pipe"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/system"
)

func Send[T any](q swarm.Queue) (chan<- T, <-chan T) {
	switch queue := q.(type) {
	case *system.Queue:
		typ := reflect.TypeOf(*new(T))
		cat := typ.Name()
		if typ.Kind() == reflect.Ptr {
			cat = typ.Elem().Name()
		}

		var ch system.MsgSendCh[T]
		ch.Msg, ch.Err = spawnSendOf[T](queue, cat)
		queue.SendCh.Attach(cat, &ch)
		return ch.Msg, ch.Err

	default:
		panic("invalid type")
	}
}

func spawnSendOf[T any](q *system.Queue, cat string) (chan T, chan T) {
	// logger.Notice("init %s sender for %s", q.id, cat)

	// TODO: configurable queue
	emit := q.Sender.Send()
	sock := make(chan T, 100)
	fail := make(chan T, 100)

	pipe.ForEach(sock, func(object T) {
		fmt.Printf("sending %v\n", object)
		msg, err := json.Marshal(object)
		fmt.Println(err)
		if err != nil {
			fail <- object
			return
		}
		emit <- mkBag(q, cat, swarm.Bytes(msg), nil, func() { fail <- object })
		// emit <- q.mkBag(cat, object, fail)
	})

	return sock, fail
}

//
//
func mkBag(q *system.Queue, cat string, msg swarm.Object, err chan<- swarm.Object, r func()) *swarm.Bag {
	return &swarm.Bag{
		Category: cat,
		System:   "swarm-example-sqs",
		Queue:    "swarm-test",
		// System:   q.sys.id,
		// Queue:    q.id,
		Object:  msg,
		StdErr:  err,
		Recover: r,
	}
}
