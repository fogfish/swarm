package system

import (
	"context"
	"sync"

	"github.com/fogfish/swarm"
)

type system struct {
	sync.Mutex
	queue map[string]*Queue

	id      string
	context context.Context
	cancel  context.CancelFunc
}

var (
	_ swarm.System = (*system)(nil)
)

/*

New creates new queueing system
*/
func NewSystem(id string) swarm.System {
	sys := &system{
		id:      id,
		queue:   make(map[string]*Queue),
		context: context.Background(),
	}

	sys.context, sys.cancel = context.WithCancel(sys.context)
	return sys
}

/*

Queue ...
*/
func (sys *system) Queue(send swarm.Sender, recv swarm.Recver) swarm.Queue {
	sys.Lock()
	defer sys.Unlock()

	queue := NewQueue(sys, send, recv)
	sys.queue[send.ID()] = queue
	return queue
}

// func (sys *system)

/*

ID return unique system ID
*/
func (sys *system) ID() string {
	return sys.id
}

/*

Listen ...
*/
func (sys *system) Listen() error {
	for _, q := range sys.queue {
		if err := q.Listen(); err != nil {
			return err
		}
	}

	return nil
}

/*

Stop ...
*/
func (sys *system) Stop() {
	for _, q := range sys.queue {
		q.Stop()
	}

	sys.cancel()
}

/*

Wait ...
*/
func (sys *system) Wait() {
	<-sys.context.Done()
}
