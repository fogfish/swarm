//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package qtest

import (
	"encoding/json"
	"testing"

	"github.com/fogfish/it"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/system"
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/bytes"
	"github.com/fogfish/swarm/queue/events"
)

const (
	Category = "note"
	Message  = "{\"some\":\"message\"}"
	Event    = ""
	Receipt  = "0x123456789abcdef"
)

type Note struct {
	Some string `json:"some"`
}

func TestEnqueue(
	t *testing.T,
	factory func(swarm.System, *swarm.Policy, chan string) (swarm.Enqueue, swarm.Dequeue),
) {
	t.Helper()

	eff := make(chan string, 1)
	sys := system.NewSystem("test-system")
	enq, deq := factory(sys, swarm.DefaultPolicy(), eff)
	q := sys.Queue("test-queue", enq, deq, swarm.DefaultPolicy())

	out, _ := queue.Enqueue[Note](q)
	bin, _ := bytes.Enqueue(q, Category+"-bin")
	evt, _ := events.Enqueue[*Note, swarm.Event[*Note]](q)
	if err := sys.Listen(); err != nil {
		panic(err)
	}

	t.Run("Success", func(t *testing.T) {
		out <- Note{Some: "message"}
		it.Ok(t).
			If(<-eff).Equal(Message)
	})

	t.Run("Success.Bytes", func(t *testing.T) {
		bin <- []byte(Message)
		it.Ok(t).
			If(<-eff).Equal(Message)
	})

	t.Run("Success.Events", func(t *testing.T) {
		evt <- &swarm.Event[*Note]{
			Object: &Note{Some: "message"},
		}
		var out swarm.Event[*Note]
		err := json.Unmarshal([]byte(<-eff), &out)

		it.Ok(t).
			IfNil(err).
			If(*out.Object).Equal(Note{Some: "message"}).
			If(string(out.Type)).Equal("note:Event[*swarm.Note").
			IfTrue(out.ID != "").
			IfTrue(out.Created != "")
	})

	// t.Run("Failure", func(t *testing.T) {
	// 	out, err := queue.Send("Some Other")
	// 	out <- swarm.Bytes(Message)

	// 	it.Ok(t).
	// 		If(<-err).Equal(swarm.Bytes(Message))
	// })

	sys.Close()
}

func TestDequeue(
	t *testing.T,
	factory func(swarm.System, *swarm.Policy, chan string) (swarm.Enqueue, swarm.Dequeue),
) {
	t.Helper()

	eff := make(chan string, 1)
	sys := system.NewSystem("test-system")
	enq, deq := factory(sys, swarm.DefaultPolicy(), eff)
	q := sys.Queue("test-queue", enq, deq, swarm.DefaultPolicy())

	msg, ack := queue.Dequeue[Note](q)
	if err := sys.Listen(); err != nil {
		panic(err)
	}

	t.Run("Success", func(t *testing.T) {
		val := <-msg
		ack <- val

		it.Ok(t).
			If(val.Object).Equal(Note{Some: "message"}).
			If(<-eff).Equal(Receipt)
	})

	sys.Close()
}
