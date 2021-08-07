package queue

import (
	"context"
	"sync"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

type tMailbox struct {
	id      swarm.Category
	channel chan swarm.Msg
}

type tRecv struct {
	build func() (<-chan *Bag, chan<- *Bag)
	value <-chan *Bag
	socks map[swarm.Category]chan swarm.Msg
	sacks map[swarm.Category]chan swarm.Msg
}

type tSend struct {
	build func() chan<- *Bag
	value chan<- *Bag
	abort <-chan *Bag
	socks map[swarm.Category]chan swarm.Msg
	fails map[swarm.Category]chan swarm.Msg
}

/*

Queue ...
*/
type Queue struct {
	sync.Mutex

	System swarm.System
	ID     string

	ctrl chan tMailbox
	acks chan<- *Bag
	recv *tRecv
	send *tSend
}

/*

New creates
*/
func New(
	sys swarm.System,
	id string,
	recv func() (<-chan *Bag, chan<- *Bag),
	send func() chan<- *Bag,
) *Queue {
	return &Queue{
		System: sys,
		ID:     id,
		ctrl:   nil,
		acks:   nil,

		recv: &tRecv{
			build: recv,
			value: nil,
			socks: make(map[swarm.Category]chan swarm.Msg),
			sacks: make(map[swarm.Category]chan swarm.Msg),
		},

		send: &tSend{
			build: send,
			value: nil,
			socks: make(map[swarm.Category]chan swarm.Msg),
			fails: make(map[swarm.Category]chan swarm.Msg),
		},
	}
}

//
//
func (q *Queue) dispatch() {
	q.ctrl = make(chan tMailbox)
	recv := q.recv.value

	q.System.Go(func(ctx context.Context) {
		logger.Notice("init %s dispatch", q.ID)
		defer close(q.ctrl)
		mailboxes := map[swarm.Category]chan swarm.Msg{}

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free %s dispatch", q.ID)
				return

			//
			case mbox := <-q.ctrl:
				mailboxes[mbox.id] = mbox.channel

			//
			case message := <-recv:
				mbox, exists := mailboxes[message.Category]
				if exists {
					mbox <- message.Object
				} else {
					logger.Notice("Category %s is not supported by queue %s ", message.Category, q.ID)
				}
			}
		}
	})
}

//
//
func (q *Queue) mkBag(cat swarm.Category, msg swarm.Msg, err chan<- swarm.Msg) *Bag {
	return &Bag{
		Target:   q.ID,
		Source:   q.System.ID(),
		Category: cat,
		Object:   msg,
		StdErr:   err,
	}
}

/*

Recv creates endpoints to receive messages and acknowledge its consumption.
*/
func (q *Queue) Recv(cat swarm.Category) (<-chan swarm.Msg, chan<- swarm.Msg) {
	q.Lock()
	defer q.Unlock()

	if q.recv.value == nil {
		q.recv.value, q.acks = q.recv.build()
		q.dispatch()
	}

	mbox, exists := q.recv.socks[cat]
	if exists {
		return mbox, q.recv.sacks[cat]
	}

	mbox = make(chan swarm.Msg)
	acks := make(chan swarm.Msg)

	q.System.Go(func(ctx context.Context) {
		logger.Notice("init %s recv type %s", q.ID, cat)
		defer close(mbox)
		defer close(acks)

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free %s recv type %s", q.ID, cat)
				return

			//
			case object := <-acks:
				q.acks <- q.mkBag(cat, object, nil)
			}
		}
	})

	q.ctrl <- tMailbox{id: cat, channel: mbox}
	q.recv.socks[cat] = mbox
	q.recv.sacks[cat] = acks
	return mbox, acks
}

/*

Send creates endpoints to send messages and receive errors.
*/
func (q *Queue) Send(cat swarm.Category) (chan<- swarm.Msg, <-chan swarm.Msg) {
	q.Lock()
	defer q.Unlock()

	if q.send.value == nil {
		q.send.value = q.send.build()
	}

	sock, exists := q.send.socks[cat]
	if exists {
		return sock, q.send.fails[cat]
	}

	sock = make(chan swarm.Msg)
	fail := make(chan swarm.Msg)

	q.System.Go(func(ctx context.Context) {
		logger.Notice("init %s send type %s", q.ID, cat)
		defer close(sock)
		defer close(fail)

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free %s send type %s", q.ID, cat)
				return

			//
			case object := <-sock:
				q.send.value <- q.mkBag(cat, object, fail)
			}
		}
	})

	q.send.socks[cat] = sock
	q.send.fails[cat] = fail
	return sock, fail
}
