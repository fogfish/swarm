package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

type tMailbox struct {
	id      swarm.Category
	channel chan []byte
}

type tRecv struct {
	build func() <-chan *swarm.Message
	value <-chan *swarm.Message
	socks map[swarm.Category]chan []byte
}

type tSend struct {
	build func() chan<- *swarm.Message
	value chan<- *swarm.Message
	socks map[swarm.Category]chan []byte
}

type Queue struct {
	sync.Mutex

	System swarm.System
	ctrl   chan tMailbox
	recv   *tRecv
	send   *tSend
}

//
func New(
	sys swarm.System,
	recv func() <-chan *swarm.Message,
	send func() chan<- *swarm.Message,
) *Queue {
	return &Queue{
		System: sys,
		ctrl:   nil,

		recv: &tRecv{
			build: recv,
			value: nil,
			socks: make(map[swarm.Category]chan []byte),
		},

		send: &tSend{
			build: send,
			value: nil,
			socks: make(map[swarm.Category]chan []byte),
		},
	}
	// q.dispatch(recv)
	// return q
}

//
//
func (q *Queue) dispatch() {
	q.ctrl = make(chan tMailbox)
	recv := q.recv.value

	q.System.Spawn(func(ctx context.Context) {
		logger.Notice("init %p dispatch", q)
		mailboxes := map[swarm.Category]chan []byte{}

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free %p dispatch", q)
				return

			//
			case mbox := <-q.ctrl:
				mailboxes[mbox.id] = mbox.channel

			//
			case message := <-recv:
				mbox, exists := mailboxes[message.Category]
				if !exists {
					// ..
					fmt.Printf("Unknown %v\n", message.Category)
					break
				}
				mbox <- message.Object
			}
		}
	})
}

//
func (q *Queue) Recv(cat swarm.Category) <-chan []byte {
	q.Lock()
	defer q.Unlock()

	if q.recv.value == nil {
		q.recv.value = q.recv.build()
		q.dispatch()
	}

	mbox, exists := q.recv.socks[cat]
	if exists {
		return mbox
	}

	mbox = make(chan []byte)
	q.ctrl <- tMailbox{id: cat, channel: mbox}

	q.recv.socks[cat] = mbox
	return mbox
}

//
func (q *Queue) Send(cat swarm.Category) chan<- []byte {
	q.Lock()
	defer q.Unlock()

	if q.send.value == nil {
		q.send.value = q.send.build()
	}

	sock, exists := q.send.socks[cat]
	if exists {
		return sock
	}

	sock = make(chan []byte)
	q.System.Spawn(func(ctx context.Context) {
		logger.Notice("init %p send type %s", q, cat)

		for {
			select {
			case <-ctx.Done():
				logger.Notice("free %p send type %s", q, cat)
				return
			case object := <-sock:
				q.send.value <- &swarm.Message{Category: cat, Object: object}
			}
		}
	})

	q.send.socks[cat] = sock
	return sock
}
