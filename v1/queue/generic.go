//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package queue

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/fogfish/golem/pipe"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/system"
)

/*

Enqueue creates channels to enqueue Golang structs and receive dead-letters
*/
func Enqueue[T any](q swarm.Queue, category ...string) (chan<- T, <-chan T) {
	queue := castToSystemQueue(q)
	cat := strings.ToLower(typeOf[T](category...))

	var ch system.MsgSendCh[T]
	ch.Msg, ch.Err = spawnEnqueueOf[T](queue, cat)
	queue.EnqueueCh.Attach(cat, &ch)
	return ch.Msg, ch.Err
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

/*

EnqueueSync synchronously enqueue message
*/
func EnqueueSync[T any](q swarm.Queue, msg T) error {
	queue := castToSystemQueue(q)
	cat := strings.ToLower(typeOf[T]())

	object, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return queue.Enqueue.EnqSync(
		&swarm.Bag{
			System:   queue.System.ID(),
			Queue:    queue.ID,
			Category: cat,
			Object:   object,
		},
	)
}

/*

Dequeue creates channels to receive Golang structs and send acknowledgement
*/
func Dequeue[T any](q swarm.Queue, category ...string) (<-chan *swarm.Msg[T], chan<- *swarm.Msg[T]) {
	queue := castToSystemQueue(q)
	cat := strings.ToLower(typeOf[T](category...))

	var ch system.MsgRecvCh[T]
	ch.Msg, ch.Ack = spawnDequeueOf[T](queue, cat)
	queue.DequeueCh.Attach(cat, &ch)
	return ch.Msg, ch.Ack
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
// Helpers
//

func typeOf[T any](category ...string) string {
	//
	// TODO: fix
	//   Action[*swarm.User] if container type is used
	//

	if len(category) > 0 {
		return category[0]
	}

	typ := reflect.TypeOf(*new(T))
	cat := typ.Name()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().Name()
	}

	return cat
}

func castToSystemQueue(q swarm.Queue) *system.Queue {
	switch queue := q.(type) {
	case *system.Queue:
		return queue
	default:
		logger.Critical("invalid type of queue %T", q)
	}
	return nil
}
