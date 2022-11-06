//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package system

import (
	"time"

	"github.com/fogfish/swarm"
)

type Closer interface {
	Sync()
	Close()
}

/*

MsgSendCh is the pair of channel, exposed by the queue to clients to send messages
*/
type MsgSendCh[T any] struct {
	Msg chan T // channel to send message out
	Err chan T // channel to recv failed messages
}

func (ch *MsgSendCh[T]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Err) == 0 {
			break
		}
	}
}

func (ch *MsgSendCh[T]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Err)
}

/*

msgRecv is the pair of channel, exposed by the queue to clients to recv messages
*/
type MsgRecvCh[T any] struct {
	Msg chan *swarm.Msg[T] // channel to recv message
	Ack chan *swarm.Msg[T] // channel to send acknowledgement
}

func (ch *MsgRecvCh[T]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Ack) == 0 {
			break
		}
	}
}

func (ch *MsgRecvCh[T]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Ack)
}

/*

MsgSendCh is the pair of channel, exposed by the queue to clients to send messages
*/
type EvtSendCh[T any, E swarm.EventKind[T]] struct {
	Msg chan *E // channel to send message out
	Err chan *E // channel to recv failed messages
}

func (ch *EvtSendCh[T, E]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Err) == 0 {
			break
		}
	}
}

func (ch *EvtSendCh[T, E]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Err)
}

/*

msgRecv is the pair of channel, exposed by the queue to clients to recv messages
*/
type EvtRecvCh[T any, E swarm.EventKind[T]] struct {
	Msg chan *E // channel to recv message
	Ack chan *E // channel to send acknowledgement
}

func (ch *EvtRecvCh[T, E]) Sync() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Ack) == 0 {
			break
		}
	}
}

func (ch *EvtRecvCh[T, E]) Close() {
	ch.Sync()
	close(ch.Msg)
	close(ch.Ack)
}
