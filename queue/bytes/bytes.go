//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

//
// The file declares public interface for binary queue
//

package bytes

import (
	"github.com/fogfish/golem/pipe"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/system"
)

/*

Enqueue creates channels to enqueue binary messages and receive dead-letters
*/
func Enqueue(q swarm.Queue, cat string) (chan<- []byte, <-chan []byte) {
	queue := castToSystemQueue(q)

	var ch system.MsgSendCh[[]byte]
	ch.Msg, ch.Err = spawnEnqueue(queue, cat)
	queue.EnqueueCh.Attach(cat, &ch)
	return ch.Msg, ch.Err
}

func spawnEnqueue(q *system.Queue, cat string) (chan []byte, chan []byte) {
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

/*

DequeueBytes creates channels to receive binary messages and send acknowledgement
*/
func Dequeue(q swarm.Queue, cat string) (<-chan *swarm.Msg[[]byte], chan<- *swarm.Msg[[]byte]) {
	queue := castToSystemQueue(q)

	var ch system.MsgRecvCh[[]byte]
	ch.Msg, ch.Ack = spawnDequeue(queue, cat)
	queue.DequeueCh.Attach(cat, &ch)
	return ch.Msg, ch.Ack
}

func spawnDequeue(q *system.Queue, cat string) (chan *swarm.Msg[[]byte], chan *swarm.Msg[[]byte]) {
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

//
//
//

func castToSystemQueue(q swarm.Queue) *system.Queue {
	switch queue := q.(type) {
	case *system.Queue:
		return queue
	default:
		logger.Critical("invalid type of queue %T", q)
	}
	return nil
}
