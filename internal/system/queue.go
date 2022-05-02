package system

import (
	"context"
	"sync"
	"time"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

type tMailbox struct {
	id      swarm.Category
	channel chan swarm.MsgV0
}

/*

msgSend is the pair of channel, exposed by the queue to clients to send messages
*/
type msgSend struct {
	msg chan swarm.MsgV0 // channel to send message out
	err chan swarm.MsgV0 // channel to recv failed messages
}

/*

msgRecv is the pair of channel, exposed by the queue to clients to recv messages
*/
type msgRecv struct {
	msg chan swarm.MsgV0 // channel to recv message
	ack chan swarm.MsgV0 // channel to send acknowledgement
}

/*

Queue ...
*/
type Queue struct {
	sync.Mutex
	sys *system
	id  string

	ctrl chan tMailbox

	queue swarm.EventBus
	qSend chan<- *swarm.Bag
	qRecv <-chan *swarm.Bag
	qConf chan<- *swarm.Bag

	recv map[swarm.Category]msgRecv
	send map[swarm.Category]msgSend
}

func NewQueue(sys *system, queue swarm.EventBus) *Queue {
	return &Queue{
		sys:   sys,
		id:    queue.ID(),
		ctrl:  make(chan tMailbox, 10000),
		queue: queue,
		recv:  make(map[swarm.Category]msgRecv),
		send:  make(map[swarm.Category]msgSend),
	}
}

//
//
func (q *Queue) dispatch() {
	q.sys.Go(func(ctx context.Context) {
		logger.Notice("init %s dispatch", q.id)
		defer close(q.ctrl)
		mailboxes := map[swarm.Category]chan swarm.MsgV0{}

		// TODO: this is a temporary solution to pass tests
		//       mbox <- message.Object fails when system is shutdown
		defer func() {
			if err := recover(); err != nil {
				logger.Error("fail %s dispatch: %v", q.id, err)
			}
		}()

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free %s dispatch", q.id)
				return

			//
			case mbox := <-q.ctrl:
				mailboxes[mbox.id] = mbox.channel

			//
			case message := <-q.qRecv:
				mbox, exists := mailboxes[message.Category]
				if exists {
					// TODO: blocked until actor consumes it
					//       it prevents proper clean-up strategy
					mbox <- message.Object
				} else {
					// TODO: at boot time category handler might not be registered yet
					logger.Notice("Category %s is not supported by queue %s ", message.Category, q.id)
				}
			}
		}
	})
}

//
//
func (q *Queue) mkBag(cat swarm.Category, msg swarm.MsgV0, err chan<- swarm.MsgV0) *swarm.Bag {
	return &swarm.Bag{
		Target:   q.id,     // subject
		Source:   q.sys.id, // namespace
		Category: cat,      // type
		Object:   msg,
		StdErr:   err,
	}
}

/*

Recv creates endpoints to receive messages and acknowledge its consumption.

Note: singleton is required for scalability
*/
func (q *Queue) Recv(cat swarm.Category) (<-chan swarm.MsgV0, chan<- swarm.MsgV0) {
	q.Lock()
	defer q.Unlock()

	if ch, exists := q.recv[cat]; exists {
		return ch.msg, ch.ack
	}

	mbox, acks := spawnRecvOf(q, cat)
	q.recv[cat] = msgRecv{msg: mbox, ack: acks}
	q.ctrl <- tMailbox{id: cat, channel: mbox}
	return mbox, acks
}

/*

spawnRecvTypeOf creates a dedicated go routine to proxy "typed" messages to queue
*/
func spawnRecvOf(q *Queue, cat swarm.Category) (chan swarm.MsgV0, chan swarm.MsgV0) {
	mbox := make(chan swarm.MsgV0)
	acks := make(chan swarm.MsgV0)

	q.sys.Go(func(ctx context.Context) {
		logger.Notice("init %s receiver for %s", q.id, cat)
		defer close(mbox)
		defer close(acks)

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free %s receiver for %s", q.id, cat)
				return

			//
			case object := <-acks:
				q.qConf <- q.mkBag(cat, object, nil)
			}
		}
	})

	return mbox, acks
}

/*

Send creates endpoints to send messages and receive errors.
*/
func (q *Queue) Send(cat swarm.Category) (chan<- swarm.MsgV0, <-chan swarm.MsgV0) {
	q.Lock()
	defer q.Unlock()

	if ch, exists := q.send[cat]; exists {
		return ch.msg, ch.err
	}

	sock, fail := spawnSendOf(q, cat)
	q.send[cat] = msgSend{msg: sock, err: fail}
	return sock, fail
}

func spawnSendOf(q *Queue, cat swarm.Category) (chan swarm.MsgV0, chan swarm.MsgV0) {
	// TODO: configurable queue
	sock := make(chan swarm.MsgV0, 100)
	fail := make(chan swarm.MsgV0, 100)

	q.sys.Go(func(ctx context.Context) {
		logger.Notice("init %s sender for %s", q.id, cat)
		defer close(sock)
		defer close(fail)

		// TODO: this is a temporary solution to pass tests
		//       mbox <- message.Object fails when system is shutdown
		defer func() {
			if err := recover(); err != nil {
				logger.Error("queue %s sender for %s failed %v", q.id, cat, err)
			}
		}()

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free %s sender for %s", q.id, cat)
				return

			//
			case object := <-sock:
				q.qSend <- q.mkBag(cat, object, fail)
			}
		}
	})

	return sock, fail
}

// Wait activates queue transport protocol
func (q *Queue) Listen() error {
	if len(q.send) > 0 {
		ch, err := q.queue.Send()
		if err != nil {
			return err
		}
		q.qSend = ch
	}

	// on receive, we need to listen when all consumers are connected
	// basically when sys.Wait() is called
	if len(q.recv) > 0 {
		rcv, err := q.queue.Recv()
		if err != nil {
			return err
		}
		q.qRecv = rcv

		cnf, err := q.queue.Conf()
		if err != nil {
			return err
		}
		q.qConf = cnf
		q.dispatch()
	}

	return nil
}

/*

Wait until queue idle
*/
func (q *Queue) Stop() {
	if len(q.send) > 0 {
		for {
			time.Sleep(100 * time.Millisecond)
			inflight := 0
			for _, sock := range q.send {
				inflight += len(sock.msg)
			}
			for _, fail := range q.send {
				inflight += len(fail.err)
			}
			if inflight == 0 {
				break
			}
		}

		// emit control message to ensure that queue is idle
		ctrl := make(chan swarm.MsgV0)
		q.qSend <- q.mkBag("", swarm.Bytes("+++"), ctrl)
		<-ctrl
	}
}
