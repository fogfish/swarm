package system

import (
	"sync"
	"time"

	"github.com/fogfish/golem/pipe"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

type tMailbox struct {
	id      string
	channel chan swarm.Object
}

/*

msgSend is the pair of channel, exposed by the queue to clients to send messages
*/
type msgSend struct {
	msg chan swarm.Object // channel to send message out
	err chan swarm.Object // channel to recv failed messages
}

/*

msgRecv is the pair of channel, exposed by the queue to clients to recv messages
*/
type msgRecv struct {
	msg chan swarm.Object // channel to recv message
	ack chan swarm.Object // channel to send acknowledgement
}

/*

Queue ...
*/
type Queue struct {
	sync.Mutex
	sys *system
	id  string

	ctrl chan tMailbox

	qrecv swarm.Recver
	recv  map[string]msgRecv

	qsend swarm.Sender
	send  map[string]msgSend
}

func NewQueue(
	sys *system,
	qsend swarm.Sender,
	qrecv swarm.Recver,
	// policy *swarm.Policy,
) *Queue {
	return &Queue{
		sys: sys,
		id:  qsend.ID(),

		ctrl: make(chan tMailbox, 10000),

		qrecv: qrecv,
		recv:  make(map[string]msgRecv),

		qsend: qsend,
		send:  make(map[string]msgSend),
	}
}

//
//
func (q *Queue) dispatch() {
	go func() {
		logger.Notice("init %s dispatch", q.id)
		mailboxes := map[string]chan swarm.Object{}
		sock := q.qrecv.Recv()

		// Note: this is required to gracefull stop dispatcher when channel is closed
		defer func() {
			if err := recover(); err != nil {
			}
			logger.Notice("free %s dispatch", q.id)
		}()

		for {
			select {
			//
			case mbox, ok := <-q.ctrl:
				if !ok {
					return
				}
				mailboxes[mbox.id] = mbox.channel

			//
			case message, ok := <-sock:
				if !ok {
					return
				}

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
	}()
}

//
//
func (q *Queue) mkBag(cat string, msg swarm.Object, err chan<- swarm.Object) *swarm.Bag {
	return &swarm.Bag{
		Category: cat,
		System:   q.sys.id,
		Queue:    q.id,
		Object:   msg,
		StdErr:   err,
	}
}

/*

Recv creates endpoints to receive messages and acknowledge its consumption.

Note: singleton is required for scalability
*/
func (q *Queue) Recv(cat string) (<-chan swarm.Object, chan<- swarm.Object) {
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
func spawnRecvOf(q *Queue, cat string) (chan swarm.Object, chan swarm.Object) {
	logger.Notice("init %s receiver for %s", q.id, cat)

	conf := q.qrecv.Conf()
	mbox := make(chan swarm.Object)
	acks := make(chan swarm.Object)

	pipe.ForEach(acks, func(object swarm.Object) {
		conf <- q.mkBag(cat, object, nil)
	})

	return mbox, acks
}

/*

Send creates endpoints to send messages and receive errors.
*/
func (q *Queue) Send(cat string) (chan<- swarm.Object, <-chan swarm.Object) {
	q.Lock()
	defer q.Unlock()

	if ch, exists := q.send[cat]; exists {
		return ch.msg, ch.err
	}

	sock, fail := spawnSendOf(q, cat)
	q.send[cat] = msgSend{msg: sock, err: fail}
	return sock, fail
}

func spawnSendOf(q *Queue, cat string) (chan swarm.Object, chan swarm.Object) {
	logger.Notice("init %s sender for %s", q.id, cat)

	// TODO: configurable queue
	emit := q.qsend.Send()
	sock := make(chan swarm.Object, 100)
	fail := make(chan swarm.Object, 100)

	pipe.ForEach(sock, func(object swarm.Object) {
		emit <- q.mkBag(cat, object, fail)
	})

	return sock, fail
}

// Wait activates queue transport protocol
func (q *Queue) Listen() error {
	if q.send != nil && len(q.send) > 0 {
		if err := q.qsend.Start(); err != nil {
			return err
		}
	}

	if q.recv != nil && len(q.recv) > 0 {
		if err := q.qrecv.Start(); err != nil {
			return err
		}
		q.dispatch()
	}

	return nil
}

/*

Wait until queue idle
*/
func (q *Queue) Stop() {
	if q.send != nil && len(q.send) > 0 {
		q.sendWaitIdle()
		q.qsend.Close()
		q.sendClose()
	}

	if q.recv != nil && len(q.recv) > 0 {
		q.qrecv.Close()
		q.recvClose()
	}
}

func (q *Queue) sendWaitIdle() {
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
	ctrl := make(chan swarm.Object)
	q.qsend.Send() <- q.mkBag("", swarm.Bytes("+++"), ctrl)
	<-ctrl
}

func (q *Queue) sendClose() {
	for _, v := range q.send {
		close(v.msg)
		close(v.err)
	}
	q.qsend = nil
}

func (q *Queue) recvClose() {
	close(q.ctrl)

	for _, v := range q.recv {
		close(v.msg)
		close(v.ack)
	}
	q.recv = nil
}
