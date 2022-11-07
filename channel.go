package swarm

import (
	"sync"
	"time"
)

// Channel is abstract concept of channel(s)
type Channel interface {
	Sync()
	Close()
}

/*

MsgEnqCh is the pair of channel, exposed by the queue for enqueuing the messages
*/
type MsgEnqCh[T any] struct {
	Msg chan T // channel to send message out
	Err chan T // channel to recv failed messages
}

func NewMsgEnqCh[T any]() MsgEnqCh[T] {
	return MsgEnqCh[T]{
		Msg: make(chan T),
		Err: make(chan T),
	}
}

func (ch MsgEnqCh[T]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Err) == 0 {
			break
		}
	}
}

func (ch MsgEnqCh[T]) Close() {
	ch.Sync()

	close(ch.Msg)
	close(ch.Err)
}

/*

msgRecv is the pair of channel, exposed by the queue to clients to recv messages
*/
type MsgDeqCh[T any] struct {
	Msg chan *Msg[T] // channel to recv message
	Ack chan *Msg[T] // channel to send acknowledgement
}

func NewMsgDeqCh[T any]() MsgDeqCh[T] {
	return MsgDeqCh[T]{
		Msg: make(chan *Msg[T]),
		Ack: make(chan *Msg[T]),
	}
}

func (ch MsgDeqCh[T]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Ack) == 0 {
			break
		}
	}
}

func (ch MsgDeqCh[T]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Ack)
}

/*

EvtEnqCh is the pair of channel, exposed by the queue to clients to send messages
*/
type EvtEnqCh[T any, E EventKind[T]] struct {
	Msg chan *E // channel to send message out
	Err chan *E // channel to recv failed messages
}

func NewEvtEnqCh[T any, E EventKind[T]]() EvtEnqCh[T, E] {
	return EvtEnqCh[T, E]{
		Msg: make(chan *E),
		Err: make(chan *E),
	}
}

func (ch EvtEnqCh[T, E]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Err) == 0 {
			break
		}
	}
}

func (ch EvtEnqCh[T, E]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Err)
}

/*

msgRecv is the pair of channel, exposed by the queue to clients to recv messages
*/
type EvtDeqCh[T any, E EventKind[T]] struct {
	Msg chan *E // channel to recv message
	Ack chan *E // channel to send acknowledgement
}

func NewEvtDeqCh[T any, E EventKind[T]]() EvtDeqCh[T, E] {
	return EvtDeqCh[T, E]{
		Msg: make(chan *E),
		Ack: make(chan *E),
	}
}

func (ch EvtDeqCh[T, E]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Ack) == 0 {
			break
		}
	}
}

func (ch EvtDeqCh[T, E]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Ack)
}

/*

Channels
*/
type Channels struct {
	sync.Mutex
	channels map[string]Channel
}

func NewChannels() *Channels {
	return &Channels{
		channels: make(map[string]Channel),
	}
}

func (chs *Channels) Length() int {
	return len(chs.channels)
}

func (chs *Channels) Attach(id string, ch Channel) {
	chs.Lock()
	defer chs.Unlock()

	chs.channels[id] = ch
}

func (chs *Channels) Sync() {
	for _, ch := range chs.channels {
		ch.Sync()
	}
}

func (chs *Channels) Close() {
	chs.Lock()
	defer chs.Unlock()

	for _, ch := range chs.channels {
		ch.Close()
	}

	chs.channels = make(map[string]Channel)
}
