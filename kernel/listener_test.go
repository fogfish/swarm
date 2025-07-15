//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import (
	"testing"
	"time"

	"github.com/fogfish/golem/optics"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
)

func TestListenerFactory(t *testing.T) {
	conf := newConfig()
	mock := mockFactory{}

	t.Run("Kernel", func(t *testing.T) {
		k := New(nil, NewListener(newMockCathode(nil, nil), conf.kernel))
		go func() {
			time.Sleep(conf.yieldBeforeClose)
			k.Close()
		}()
		k.Await()
	})

	t.Run("None", func(t *testing.T) {
		k := NewListener(mock.ListenerCore(nil, nil), conf.kernel)
		go k.Await()
		k.Close()
	})
}

func TestRecvChan(t *testing.T) {
	recvTest(t,
		encoding.ForTyped[string](),
		RecvChan,
		optics.ForProduct1[swarm.Msg[string], any]("IOContext"),
		[]swarm.Bag{
			{
				Category:  "string",
				Digest:    "1",
				IOContext: "context",
				Object:    []byte(`"1"`),
			},
		},
	)
}

func TestRecvEvent(t *testing.T) {
	type E = swarm.Event[swarm.Meta, string]

	recvTest(t,
		encoding.ForEvent[E]("testReal", "testAgent"),
		RecvEvent,
		optics.ForProduct1[E, any]("IOContext"),
		[]swarm.Bag{
			{
				Category:  "string",
				Digest:    "1",
				IOContext: "context",
				Object:    []byte(`{"meta":{"type": "string"}, "data": "1"}`),
			},
		},
	)
}

func recvTest[M any, T any](
	t *testing.T,
	codec Decoder[T],
	recvChan func(k *ListenerIO, codec Decoder[T]) (<-chan M, chan<- M),
	ioContext optics.Lens[M, any],
	bag []swarm.Bag,
) {
	t.Helper()
	conf := newConfig()

	mock := mockFactory{}
	none := mock.ListenerCore(nil, nil)
	pass := mock.ListenerCore(make(chan string), bag)

	t.Run("None", func(t *testing.T) {
		k := NewListener(none, conf.kernel)

		recvChan(k, codec)

		go k.Await()
		k.Close()
	})

	t.Run("Dequeue.1", func(t *testing.T) {
		k := NewListener(pass, conf.kernel)
		rcv, ack := recvChan(k, codec)
		go k.Await()

		ack <- <-rcv
		it.Then(t).Should(
			it.Equal(string(<-pass.ack), `1`),
		)

		k.Close()
	})

	t.Run("Dequeue.1.Context", func(t *testing.T) {
		k := NewListener(pass, conf.kernel)
		rcv, ack := recvChan(k, codec)
		go k.Await()

		msg := <-rcv
		ack <- msg
		it.Then(t).Should(
			it.Equal(string(<-pass.ack), `1`),
			it.Equal(ioContext.Get(&msg).(string), "context"),
		)

		k.Close()
	})

	t.Run("Dequeue.N.Backlog", func(t *testing.T) {
		cfg := swarm.NewConfig()
		cfg.CapAck = 4
		cfg.PollFrequency = 1 * time.Millisecond
		k := NewListener(pass, cfg)
		rcv, ack := recvChan(k, codec)
		go k.Await()

		ack <- <-rcv
		ack <- <-rcv
		ack <- <-rcv
		ack <- <-rcv
		go k.Close()

		it.Then(t).Should(
			it.Equal(string(<-pass.ack), `1`),
			it.Equal(string(<-pass.ack), `1`),
			it.Equal(string(<-pass.ack), `1`),
			it.Equal(string(<-pass.ack), `1`),
		)
	})

}
