//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package embedded_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/embedded"
	"github.com/fogfish/swarm/emit"
	"github.com/fogfish/swarm/listen"
)

func TestEmitter(t *testing.T) {
	t.Run("NewEmitter", func(t *testing.T) {
		q, err := embedded.Endpoint().Build()

		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Emit.Recv.1", func(t *testing.T) {
		q, err := embedded.Endpoint().Build()
		it.Then(t).Should(it.Nil(err))

		var obj string
		snd := swarm.LogDeadLetters(emit.Typed[string](q.Emitter))
		rcv, ack := listen.Typed[string](q.Listener)

		snd <- "hello world"
		go func() {
			msg := <-rcv
			obj = msg.Object
			ack <- msg

			time.Sleep(5 * time.Millisecond)
			q.Close()
		}()
		q.Await()

		it.Then(t).Should(
			it.Equal(obj, "hello world"),
		)
	})

	t.Run("Emit.Error.Recv.1", func(t *testing.T) {
		q, err := embedded.Endpoint().Build()
		it.Then(t).Should(it.Nil(err))

		var obj string
		snd := swarm.LogDeadLetters(emit.Typed[string](q.Emitter))
		rcv, ack := listen.Typed[string](q.Listener)

		snd <- "hello world"
		go func() {
			msg1 := <-rcv
			ack <- msg1.Fail(fmt.Errorf("fail"))

			msg2 := <-rcv
			obj = msg2.Object
			ack <- msg2

			time.Sleep(5 * time.Millisecond)
			q.Close()
		}()
		q.Await()

		it.Then(t).Should(
			it.Equal(obj, "hello world"),
		)
	})

}
