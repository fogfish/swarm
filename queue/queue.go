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

type Queue struct {
	sync.Mutex

	sys swarm.System

	ctrl chan tMailbox
	recv map[swarm.Category]chan []byte
	send map[swarm.Category]chan []byte
	emit chan<- *swarm.Message
}

//
func New(sys swarm.System, recv <-chan *swarm.Message, emit chan<- *swarm.Message) *Queue {
	q := &Queue{
		sys: sys,

		recv: make(map[swarm.Category]chan []byte),
		send: make(map[swarm.Category]chan []byte),
		emit: emit,
	}
	q.dispatch(recv)
	return q
}

//
//
func (q *Queue) dispatch(recv <-chan *swarm.Message) {
	q.ctrl = make(chan tMailbox)

	q.sys.Spawn(func(ctx context.Context) {
		logger.Notice("start dispatch loop %p", q)
		mailboxes := map[swarm.Category]chan []byte{}

		for {
			select {
			case <-ctx.Done():
				logger.Notice("stop dispatch loop %p", q)
				return
			case mbox := <-q.ctrl:
				mailboxes[mbox.id] = mbox.channel
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

	mbox, exists := q.recv[cat]
	if exists {
		return mbox
	}

	mbox = make(chan []byte)
	q.ctrl <- tMailbox{id: cat, channel: mbox}

	q.recv[cat] = mbox
	return mbox
}

//
func (q *Queue) Send(cat swarm.Category) chan<- []byte {
	q.Lock()
	defer q.Unlock()

	sock, exists := q.send[cat]
	if exists {
		return sock
	}

	sock = make(chan []byte)
	q.sys.Spawn(func(ctx context.Context) {
		logger.Notice("start send loop %p for %s", q, cat)

		for {
			select {
			case <-ctx.Done():
				logger.Notice("stop send loop %p for %s", q, cat)
				return
			case object := <-sock:
				q.emit <- &swarm.Message{Category: cat, Object: object}
			}
		}
	})

	q.send[cat] = sock
	return sock
}
