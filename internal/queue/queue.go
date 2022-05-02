package queue

// import (
// 	"context"
// 	"fmt"
// 	"sync"
// 	"time"

// 	"github.com/fogfish/logger"
// 	"github.com/fogfish/swarm"
// )

// type tMailbox struct {
// 	id      swarm.Category
// 	channel chan swarm.Msg
// }

// /*

// msgSend is the pair of channel, exposed by the queue to clients to send messages
// */
// type msgSend struct {
// 	msg chan swarm.Msg // channel to send message out
// 	err chan swarm.Msg // channel to recv failed messages
// }

// /*

// msgRecv is the pair of channel, exposed by the queue to clients to recv messages
// */
// type msgRecv struct {
// 	msg chan swarm.Msg // channel to recv message
// 	ack chan swarm.Msg // channel to send acknowledgement
// }

// type tRecv struct {
// 	// build func() (<-chan *Bag, chan<- *Bag)
// 	// value <-chan *Bag
// 	socks map[swarm.Category]chan swarm.Msg
// 	sacks map[swarm.Category]chan swarm.Msg
// }

// type tSend struct {
// 	// build func() chan<- *Bag
// 	// value chan<- *Bag
// 	socks map[swarm.Category]chan swarm.Msg
// 	fails map[swarm.Category]chan swarm.Msg
// }

// /*

// Queue ...
// */
// type Queue struct {
// 	sync.Mutex

// 	System swarm.System
// 	ID     string

// 	wire  Wire
// 	qSend chan<- *Bag
// 	qRecv <-chan *Bag
// 	qConf chan<- *Bag

// 	ctrl chan tMailbox
// 	// acks chan<- *Bag
// 	recv *tRecv
// 	send *tSend
// }

// /*

// New creates
// */
// func New(
// 	sys swarm.System,
// 	id string,
// 	recv func() (<-chan *Bag, chan<- *Bag),
// 	send func() chan<- *Bag,
// ) *Queue {
// 	return &Queue{
// 		System: sys,
// 		ID:     id,
// 		ctrl:   nil,
// 		// acks:   nil,

// 		recv: &tRecv{
// 			// build: recv,
// 			// value: nil,
// 			socks: make(map[swarm.Category]chan swarm.Msg),
// 			sacks: make(map[swarm.Category]chan swarm.Msg),
// 		},

// 		send: &tSend{
// 			// build: send,
// 			// value: nil,
// 			socks: make(map[swarm.Category]chan swarm.Msg),
// 			fails: make(map[swarm.Category]chan swarm.Msg),
// 		},
// 	}
// }

// func NewV2(
// 	sys swarm.System,
// 	id string,
// 	wire Wire,
// ) *Queue {
// 	q := &Queue{
// 		System: sys,
// 		ID:     id,
// 		wire:   wire,
// 		ctrl:   nil,
// 		// acks:   nil,

// 		recv: &tRecv{
// 			socks: make(map[swarm.Category]chan swarm.Msg),
// 			sacks: make(map[swarm.Category]chan swarm.Msg),
// 		},

// 		send: &tSend{
// 			socks: make(map[swarm.Category]chan swarm.Msg),
// 			fails: make(map[swarm.Category]chan swarm.Msg),
// 		},
// 	}
// 	return q
// }

// //
// //
// func (q *Queue) Sys() swarm.System {
// 	return q.System
// }

// // Listen spawns queue transport
// func (q *Queue) Listen() error {
// 	if len(q.send.socks) > 0 {
// 		q.qSend = q.wire.Send()
// 	}

// 	if len(q.recv.socks) > 0 {
// 		q.qRecv = q.wire.Recv()
// 		q.qConf = q.wire.Conf()
// 		q.dispatch()
// 	}

// 	return nil
// }

// //
// //
// func (q *Queue) dispatch() {
// 	q.ctrl = make(chan tMailbox)

// 	q.System.Go(func(ctx context.Context) {
// 		logger.Notice("init %s dispatch", q.ID)
// 		defer close(q.ctrl)
// 		mailboxes := map[swarm.Category]chan swarm.Msg{}

// 		// TODO: this is a temporary solution to pass tests
// 		//       mbox <- message.Object fails when system is shutdown
// 		defer func() {
// 			if err := recover(); err != nil {
// 				logger.Error("fail %s dispatch: %v", q.ID, err)
// 			}
// 		}()

// 		for {
// 			select {
// 			//
// 			case <-ctx.Done():
// 				logger.Notice("free %s dispatch", q.ID)
// 				return

// 			//
// 			case mbox := <-q.ctrl:
// 				mailboxes[mbox.id] = mbox.channel

// 			//
// 			case message := <-q.qRecv:
// 				mbox, exists := q.recv.socks[message.Category] //  mailboxes[message.Category]
// 				if exists {
// 					// TODO: blocked until actor consumes it
// 					//       it prevents proper clean-up strategy
// 					mbox <- message.Object
// 				} else {
// 					// TODO: at boot time category handler might not be registered yet
// 					logger.Notice("Category %s is not supported by queue %s ", message.Category, q.ID)
// 				}
// 			}
// 		}
// 	})
// }

// //
// //
// func (q *Queue) mkBag(cat swarm.Category, msg swarm.Msg, err chan<- swarm.Msg) *Bag {
// 	return &Bag{
// 		Target:   q.ID,
// 		Source:   q.System.ID(),
// 		Category: cat,
// 		Object:   msg,
// 		StdErr:   err,
// 	}
// }

// /*

// Recv creates endpoints to receive messages and acknowledge its consumption.

// Note: singleton is required for scalability
// */
// func (q *Queue) Recv(cat swarm.Category) (<-chan swarm.Msg, chan<- swarm.Msg) {
// 	q.Lock()
// 	defer q.Unlock()

// 	mbox, exists := q.recv.socks[cat]
// 	if exists {
// 		return mbox, q.recv.sacks[cat]
// 	}

// 	mbox, acks := spawnRecvTypeOf(q, cat)

// 	q.ctrl <- tMailbox{id: cat, channel: mbox}
// 	q.recv.socks[cat] = mbox
// 	q.recv.sacks[cat] = acks
// 	fmt.Println(q.recv.socks)

// 	return mbox, acks
// }

// /*

// spawnRecvTypeOf creates a dedicated go routine to proxy "typed" messages to queue
// */
// func spawnRecvTypeOf(q *Queue, cat swarm.Category) (chan swarm.Msg, chan swarm.Msg) {
// 	mbox := make(chan swarm.Msg)
// 	acks := make(chan swarm.Msg)

// 	q.System.Go(func(ctx context.Context) {
// 		logger.Notice("init %s recv type %s", q.ID, cat)
// 		defer close(mbox)
// 		defer close(acks)

// 		for {
// 			select {
// 			//
// 			case <-ctx.Done():
// 				logger.Notice("free %s recv type %s", q.ID, cat)
// 				return

// 			//
// 			case object := <-acks:
// 				q.qConf <- q.mkBag(cat, object, nil)
// 			}
// 		}
// 	})

// 	return mbox, acks
// }

// // func Send[T any](q *Queue) (chan<- T, <-chan T) {
// // 	cat := reflect.TypeOf(*new(T)).Name()
// // 	fmt.Println(cat)

// // 	send, fail := createSend[T]()
// // 	return send, fail
// // }

// // func createSend[T any]() (chan<- T, <-chan T) {
// // 	// How to handle fail for typed message
// // 	// - make proxy for fail channel with decoding of message

// // 	return nil, nil
// // }

// /*

// Send creates endpoints to send messages and receive errors.
// */
// func (q *Queue) Send(cat swarm.Category) (chan<- swarm.Msg, <-chan swarm.Msg) {
// 	q.Lock()
// 	defer q.Unlock()

// 	sock, exists := q.send.socks[cat]
// 	if exists {
// 		return sock, q.send.fails[cat]
// 	}

// 	// TODO: configurable queue
// 	sock = make(chan swarm.Msg, 100)
// 	fail := make(chan swarm.Msg, 100)

// 	q.System.Go(func(ctx context.Context) {
// 		logger.Notice("init %s send type %s", q.ID, cat)
// 		defer close(sock)
// 		defer close(fail)

// 		// TODO: this is a temporary solution to pass tests
// 		//       mbox <- message.Object fails when system is shutdown
// 		defer func() {
// 			if err := recover(); err != nil {
// 				logger.Error("fail %s send type %s: %v", q.ID, cat, err)
// 			}
// 		}()

// 		for {
// 			select {
// 			//
// 			case <-ctx.Done():
// 				logger.Notice("free %s send type %s", q.ID, cat)
// 				return

// 			//
// 			case object := <-sock:
// 				q.qSend <- q.mkBag(cat, object, fail)
// 			}
// 		}
// 	})

// 	q.send.socks[cat] = sock
// 	q.send.fails[cat] = fail
// 	return sock, fail
// }

// // func

// /*

// Wait until queue idle
// */
// func (q *Queue) Wait() {
// 	for {
// 		time.Sleep(100 * time.Millisecond)
// 		inflight := 0
// 		for _, sock := range q.send.socks {
// 			inflight += len(sock)
// 		}
// 		for _, fail := range q.send.fails {
// 			inflight += len(fail)
// 		}
// 		if inflight == 0 {
// 			break
// 		}
// 	}

// 	// emit control message to ensure that queue is idle
// 	ctrl := make(chan swarm.Msg)
// 	q.qSend <- q.mkBag("", swarm.Bytes("+++"), ctrl)
// 	<-ctrl
// }
