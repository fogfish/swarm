package swarm

import (
	"context"
	"sync"
)

/*

Category of message (aka subject, topic)
*/
type Category string

/*

Msg type is an abstract container for octet stream exchange via channels

  ch <- swarm.Bytes("some message")
  ...
	for msg := range ch {
		msg.Bytes()
	}
*/
type Msg interface {
	Bytes() []byte
}

/*

Bytes (octet stream) is a message to communicate via channels
*/
type Bytes []byte

/*

Bytes returns message payload (octet stream)
*/
func (b Bytes) Bytes() []byte { return b }

/*

Queue ...
*/
type Queue interface {
	/*
		Creates endpoints to receive messages and acknowledge its consumption.
	*/
	Recv(Category) (<-chan Msg, chan<- Msg)

	/*
		Creates endpoints to send messages and receive errors.
	*/
	Send(Category) (chan<- Msg, <-chan Msg)
}

/*

System ...
*/
type System interface {
	/*
		system ID
	*/
	ID() string

	/*
		spawn go routine in context of system
	*/
	Go(func(context.Context))

	/*
	 stop system and all active go routines
	*/
	Stop()
}

type system struct {
	sync.Mutex

	id      string
	context context.Context
	cancel  context.CancelFunc
}

/*

New creates new queueing system
*/
func New(id string) System {
	ctx, cancel := context.WithCancel(context.Background())

	return &system{
		id:      id,
		context: ctx,
		cancel:  cancel,
	}
}

var (
	_ System = (*system)(nil)
)

/*

 */
func (sys *system) ID() string {
	return sys.id
}

/*

Spawn ...
*/
func (sys *system) Go(f func(context.Context)) {
	go f(sys.context)
}

/*

Stop ...
*/
func (sys *system) Stop() {
	sys.cancel()
}
