//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

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

const syncInterval = 100 * time.Millisecond

/*
MsgEnqCh is the pair of channel, exposed by the queue for enqueuing the messages
*/
type MsgEnqCh[T any] struct {
	Msg  chan T // channel to send message out
	Err  chan T // channel to recv failed messages
	Busy sync.Mutex
}

func NewMsgEnqCh[T any](n int) MsgEnqCh[T] {
	return MsgEnqCh[T]{
		Msg: make(chan T, n),
		Err: make(chan T, n),
	}
}

func (ch *MsgEnqCh[T]) Sync() {
	for {
		time.Sleep(syncInterval)
		if len(ch.Msg)+len(ch.Err) == 0 {
			break
		}
	}

	time.Sleep(syncInterval)
	ch.Busy.Lock()
	defer ch.Busy.Unlock()
}

func (ch *MsgEnqCh[T]) Close() {
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

func NewMsgDeqCh[T any](n int) MsgDeqCh[T] {
	return MsgDeqCh[T]{
		Msg: make(chan *Msg[T], n),
		Ack: make(chan *Msg[T], n),
	}
}

func (ch *MsgDeqCh[T]) Sync() {
	for {
		time.Sleep(syncInterval)
		if len(ch.Msg)+len(ch.Ack) == 0 {
			break
		}
	}
}

func (ch *MsgDeqCh[T]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Ack)
}

/*
EvtEnqCh is the pair of channel, exposed by the queue to clients to send messages
*/
type EvtEnqCh[T any, E EventKind[T]] struct {
	Msg  chan *E // channel to send message out
	Err  chan *E // channel to recv failed messages
	Busy sync.Mutex
}

func NewEvtEnqCh[T any, E EventKind[T]](n int) EvtEnqCh[T, E] {
	return EvtEnqCh[T, E]{
		Msg: make(chan *E, n),
		Err: make(chan *E, n),
	}
}

func (ch *EvtEnqCh[T, E]) Sync() {
	for {
		time.Sleep(syncInterval)
		if len(ch.Msg)+len(ch.Err) == 0 {
			break
		}
	}

	time.Sleep(syncInterval)
	ch.Busy.Lock()
	defer ch.Busy.Unlock()
}

func (ch *EvtEnqCh[T, E]) Close() {
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

func NewEvtDeqCh[T any, E EventKind[T]](n int) EvtDeqCh[T, E] {
	return EvtDeqCh[T, E]{
		Msg: make(chan *E, n),
		Ack: make(chan *E, n),
	}
}

func (ch *EvtDeqCh[T, E]) Sync() {
	for {
		time.Sleep(syncInterval)
		if len(ch.Msg)+len(ch.Ack) == 0 {
			break
		}
	}
}

func (ch *EvtDeqCh[T, E]) Close() {
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
