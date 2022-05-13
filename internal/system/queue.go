package system

import (
	"sync"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

type Mailbox struct {
	ID    string
	Queue chan *swarm.Bag
}

/*

Queue ...
*/
type Queue struct {
	sync.Mutex

	System *system
	ID     string

	Policy *swarm.Policy
	Ctrl   chan Mailbox

	Enqueue   swarm.Enqueue
	EnqueueCh *Channels

	Dequeue   swarm.Dequeue
	DequeueCh *Channels

	logger logger.Logger
}

func NewQueue(
	sys *system,
	queue string,
	enq swarm.Enqueue,
	deq swarm.Dequeue,
	policy *swarm.Policy,
) *Queue {
	return &Queue{
		System: sys,
		ID:     queue,

		Policy: policy,
		Ctrl:   make(chan Mailbox, 10000),

		Enqueue:   enq,
		EnqueueCh: NewChannels(),

		Dequeue:   deq,
		DequeueCh: NewChannels(),

		logger: logger.With(
			logger.Note{
				"queue": sys.id + "://" + queue,
			},
		),
	}
}

//
//
func (q *Queue) dispatch() {
	go func() {
		q.logger.Info("init dequeue router")
		mailboxes := map[string]chan *swarm.Bag{}
		sock := q.Dequeue.Deq()

		// Note: this is required to gracefull stop dispatcher when channel is closed
		defer func() {
			if err := recover(); err != nil {
			}
			q.logger.Info("free dequeue router")
		}()

		for {
			// Note: first select prevents raise condition when reading multiple channel.
			select {
			case mbox := <-q.Ctrl:
				mailboxes[mbox.ID] = mbox.Queue
				q.logger.Debug("register category %s", mbox.ID)
				continue
			default:
			}

			select {
			case mbox, ok := <-q.Ctrl:
				if !ok {
					return
				}
				mailboxes[mbox.ID] = mbox.Queue
				q.logger.Debug("register category %s", mbox.ID)

			//
			case message, ok := <-sock:
				if !ok {
					return
				}

				mbox, exists := mailboxes[message.Category]
				if exists {
					// TODO: blocked until actor consumes it
					//       it prevents proper clean-up strategy
					mbox <- message
				} else {
					q.logger.Error("category %s is not found", message.Category)
				}
			}
		}
	}()
}

// Wait activates queue transport protocol
func (q *Queue) Listen() error {
	q.logger.Info("queue listening")

	if q.EnqueueCh.Length() > 0 {
		if err := q.Enqueue.Listen(); err != nil {
			return err
		}
	}

	if q.DequeueCh.Length() > 0 {
		if err := q.Dequeue.Listen(); err != nil {
			return err
		}
		q.dispatch()
	}

	return nil
}

/*

Wait until queue idle
*/
func (q *Queue) Close() {
	q.Lock()
	defer q.Unlock()

	if q.EnqueueCh.Length() > 0 {
		q.EnqueueCh.Close()
		q.enqueueWaitIdle()
		q.Enqueue.Close()
	}

	if q.DequeueCh.Length() > 0 {
		q.DequeueCh.Close()
		close(q.Ctrl)
		q.Dequeue.Close()
	}

	q.logger.Info("queue closed")
}

func (q *Queue) enqueueWaitIdle() {
	// emit control message to ensure that queue is idle
	ctrl := make(chan struct{})
	q.Enqueue.Enq() <- &swarm.BagStdErr{
		Bag:    swarm.Bag{Object: []byte("+++")},
		StdErr: func(error) { ctrl <- struct{}{} },
	}
	<-ctrl
}
