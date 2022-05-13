package queue

import (
	"encoding/json"
	"reflect"

	"github.com/fogfish/golem/pipe"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/system"
)

func typeOf[T any]() string {
	typ := reflect.TypeOf(*new(T))
	cat := typ.Name()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().Name()
	}

	return cat
}

//
//
func Enqueue[T any](q swarm.Queue) (chan<- T, <-chan T) {
	switch queue := q.(type) {
	case *system.Queue:
		cat := typeOf[T]()

		var ch system.MsgSendCh[T]
		ch.Msg, ch.Err = spawnEnqueueOf[T](queue, cat)
		queue.EnqueueCh.Attach(cat, &ch)
		return ch.Msg, ch.Err

	default:
		logger.Critical("invalid type of queue %T", q)
	}

	return nil, nil
}

func spawnEnqueueOf[T any](q *system.Queue, cat string) (chan T, chan T) {
	logger := logger.With(
		logger.Note{
			"queue": q.System.ID() + "://" + q.ID + "/" + cat,
		},
	)
	logger.Info("init enqueue")

	emit := q.Enqueue.Enq()
	sock := make(chan T, q.Policy.QueueCapacity)
	fail := make(chan T, q.Policy.QueueCapacity)

	pipe.ForEach(sock, func(object T) {
		msg, err := json.Marshal(object)
		if err != nil {
			fail <- object
			return
		}
		emit <- &swarm.BagStdErr{
			Bag: swarm.Bag{
				System:   q.System.ID(),
				Queue:    q.ID,
				Category: cat,
				Object:   msg,
			},
			StdErr: func(error) { fail <- object },
		}
	})

	return sock, fail
}

//
//
func EnqueueBytes(q swarm.Queue, cat string) (chan<- []byte, <-chan []byte) {
	switch queue := q.(type) {
	case *system.Queue:
		var ch system.MsgSendCh[[]byte]
		ch.Msg, ch.Err = spawnEnqueueBytesOf(queue, cat)
		queue.EnqueueCh.Attach(cat, &ch)
		return ch.Msg, ch.Err

	default:
		logger.Critical("invalid type of queue %T", q)
	}

	return nil, nil
}

func spawnEnqueueBytesOf(q *system.Queue, cat string) (chan []byte, chan []byte) {
	logger := logger.With(
		logger.Note{
			"queue": q.System.ID() + "://" + q.ID + "/" + cat,
		},
	)
	logger.Info("init enqueue for bytes")

	emit := q.Enqueue.Enq()
	sock := make(chan []byte, q.Policy.QueueCapacity)
	fail := make(chan []byte, q.Policy.QueueCapacity)

	pipe.ForEach(sock, func(object []byte) {
		emit <- &swarm.BagStdErr{
			Bag: swarm.Bag{
				System:   q.System.ID(),
				Queue:    q.ID,
				Category: cat,
				Object:   object,
			},
			StdErr: func(error) { fail <- object },
		}
	})

	return sock, fail
}

//
//
func Dequeue[T any](q swarm.Queue) (<-chan *swarm.Msg[T], chan<- *swarm.Msg[T]) {
	switch queue := q.(type) {
	case *system.Queue:
		cat := typeOf[T]()

		var ch system.MsgRecvCh[T]
		ch.Msg, ch.Ack = spawnDequeueOf[T](queue, cat)
		queue.DequeueCh.Attach(cat, &ch)
		return ch.Msg, ch.Ack
	default:
		logger.Critical("invalid type of queue %T", q)
	}
	return nil, nil
}

func spawnDequeueOf[T any](q *system.Queue, cat string) (chan *swarm.Msg[T], chan *swarm.Msg[T]) {
	logger := logger.With(
		logger.Note{
			"queue": q.System.ID() + "://" + q.ID + "/" + cat,
		},
	)
	logger.Info("init dequeue")

	conf := q.Dequeue.Ack()
	mbox := make(chan *swarm.Bag)
	acks := make(chan *swarm.Msg[T])

	pipe.ForEach(acks, func(object *swarm.Msg[T]) {
		conf <- &swarm.Bag{
			Category: cat,
			System:   q.System.ID(),
			Queue:    q.ID,
			Digest:   object.Digest,
		}
	})

	recv := pipe.MaybeMap(mbox, func(bag *swarm.Bag) (*swarm.Msg[T], error) {
		msg := &swarm.Msg[T]{Digest: bag.Digest}
		if err := json.Unmarshal(bag.Object, &msg.Object); err != nil {
			logger.Error("failed to dequeue %v", err)
			return nil, err
		}

		return msg, nil
	})

	q.Ctrl <- system.Mailbox{ID: cat, Queue: mbox}
	return recv, acks
}

//
//
func DequeueBytes(q swarm.Queue, cat string) (<-chan *swarm.Msg[[]byte], chan<- *swarm.Msg[[]byte]) {
	switch queue := q.(type) {
	case *system.Queue:
		var ch system.MsgRecvCh[[]byte]
		ch.Msg, ch.Ack = spawnDequeueBytes(queue, cat)
		queue.DequeueCh.Attach(cat, &ch)
		return ch.Msg, ch.Ack
	default:
		logger.Critical("invalid type of queue %T", q)
	}
	return nil, nil
}

func spawnDequeueBytes(q *system.Queue, cat string) (chan *swarm.Msg[[]byte], chan *swarm.Msg[[]byte]) {
	logger := logger.With(
		logger.Note{
			"queue": q.System.ID() + "://" + q.ID + "/" + cat,
		},
	)
	logger.Info("init dequeue for bytes")

	conf := q.Dequeue.Ack()
	mbox := make(chan *swarm.Bag)
	acks := make(chan *swarm.Msg[[]byte])

	pipe.ForEach(acks, func(object *swarm.Msg[[]byte]) {
		conf <- &swarm.Bag{
			Category: cat,
			System:   q.System.ID(),
			Queue:    q.ID,
			Digest:   object.Digest,
		}
	})

	recv := pipe.Map(mbox, func(bag *swarm.Bag) *swarm.Msg[[]byte] {
		return &swarm.Msg[[]byte]{
			Object: bag.Object,
			Digest: bag.Digest,
		}
	})

	q.Ctrl <- system.Mailbox{ID: cat, Queue: mbox}
	return recv, acks
}
