//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

//
// The file declares public interface for swarm.Event queue
//

package events

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/fogfish/golem/pipe"
	"github.com/fogfish/guid"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/system"
)

/*


Enqueue creates channels to enqueue Golang structs and receive dead-letters
*/
func Enqueue[T any, E swarm.EventKind[T]](q swarm.Queue) (chan<- *E, <-chan *E) {
	queue := castToSystemQueue(q)
	cat := strings.ToLower(typeOf[T]()) + ":" + typeOf[E]()

	var ch system.EvtSendCh[T, E]
	ch.Msg, ch.Err = spawnEnqueueOf[T, E](queue, cat)
	queue.EnqueueCh.Attach(cat, &ch)
	return ch.Msg, ch.Err
}

func spawnEnqueueOf[T any, E swarm.EventKind[T]](q *system.Queue, cat string) (chan *E, chan *E) {
	logger := logger.With(
		logger.Note{
			"queue": q.System.ID() + "://" + q.ID + "/" + cat,
		},
	)
	logger.Info("init enqueue for events")

	//
	// building memory layout to make unsafe struct modification
	kindT := swarm.Event[T]{}
	offID, offType, offCreated :=
		unsafe.Offsetof(kindT.ID),
		unsafe.Offsetof(kindT.Type),
		unsafe.Offsetof(kindT.Created)

	emit := q.Enqueue.Enq()
	sock := make(chan *E, q.Policy.QueueCapacity)
	fail := make(chan *E, q.Policy.QueueCapacity)

	pipe.ForEach(sock, func(object *E) {
		evt := unsafe.Pointer(object)

		// patch ID
		if len(*(*string)(unsafe.Pointer(uintptr(evt) + offID))) == 0 {
			id := guid.G.K(q.Clock).String()
			*(*string)(unsafe.Pointer(uintptr(evt) + offID)) = id
		}

		// patch Type
		if len(*(*string)(unsafe.Pointer(uintptr(evt) + offType))) == 0 {
			*(*string)(unsafe.Pointer(uintptr(evt) + offType)) = cat
		}

		// patch Created
		if len(*(*string)(unsafe.Pointer(uintptr(evt) + offCreated))) == 0 {
			t := time.Now().Format(time.RFC3339)
			*(*string)(unsafe.Pointer(uintptr(evt) + offCreated)) = t
		}

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

Dequeue creates channels to receive Golang structs and send acknowledgement
*/
func Dequeue[T any, E swarm.EventKind[T]](q swarm.Queue) (<-chan *E, chan<- *E) {
	queue := castToSystemQueue(q)
	cat := strings.ToLower(typeOf[T]()) + ":" + typeOf[E]()

	var ch system.EvtRecvCh[T, E]
	ch.Msg, ch.Ack = spawnDequeueOf[T, E](queue, cat)
	queue.DequeueCh.Attach(cat, &ch)
	return ch.Msg, ch.Ack
}

func spawnDequeueOf[T any, E swarm.EventKind[T]](q *system.Queue, cat string) (chan *E, chan *E) {
	logger := logger.With(
		logger.Note{
			"queue": q.System.ID() + "://" + q.ID + "/" + cat,
		},
	)
	logger.Info("init dequeue")

	//
	// building memory layout to make unsafe struct reading
	kindT := swarm.Event[T]{}
	offDigest := unsafe.Offsetof(kindT.Digest)

	conf := q.Dequeue.Ack()
	mbox := make(chan *swarm.Bag)
	acks := make(chan *E)

	pipe.ForEach(acks, func(object *E) {
		evt := unsafe.Pointer(object)
		digest := *(*string)(unsafe.Pointer(uintptr(evt) + offDigest))

		conf <- &swarm.Bag{
			Category: cat,
			System:   q.System.ID(),
			Queue:    q.ID,
			Digest:   digest,
		}
	})

	recv := pipe.MaybeMap(mbox, func(bag *swarm.Bag) (*E, error) {
		evt := new(E)
		if err := json.Unmarshal(bag.Object, evt); err != nil {
			logger.Error("failed to dequeue %v", err)
			return nil, err
		}

		ptr := unsafe.Pointer(evt)
		*(*string)(unsafe.Pointer(uintptr(ptr) + offDigest)) = bag.Digest

		return evt, nil
	})

	q.Ctrl <- system.Mailbox{ID: cat, Queue: mbox}
	return recv, acks
}

//
//
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
