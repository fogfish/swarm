package swarm

import (
	"context"
	"encoding/json"
	"sync"
)

/*

Queue ...

*/
type Queue interface {
	Recv(Category) <-chan []byte
	Send(Category) chan<- []byte
}

/*

Category of message
*/
type Category string

/*

Message attributes
 - Target (Queue name)
 - Source (Actor name, mailbox)
 - Category (DetailType)
 - Object (Detail)

swarm.Enfold(interface{}).WithXXX(...).Send(Queue)
queue <- swarm.Enfold(interface{}).WithXXX(...)

sys.Connect("target").Send("category") <- []byte("xxx")


*/
type Message struct {
	Category Category
	Object   json.RawMessage
}

/*

System ...
*/
type System interface {
	// spawn go routine in context of system
	Spawn(func(context.Context))
	// stop system and all active go routines
	Stop()
}

type system struct {
	sync.Mutex

	context context.Context
	cancel  context.CancelFunc
}

// TODO: system name for default source
func New(name string) System {
	ctx, cancel := context.WithCancel(context.Background())

	return &system{
		context: ctx,
		cancel:  cancel,
	}
}

var (
	_ System = (*system)(nil)
)

/*

Spawn ...
*/
func (sys *system) Spawn(f func(context.Context)) {
	go f(sys.context)
}

/*

Stop ...
*/
func (sys *system) Stop() {
	sys.cancel()
}
