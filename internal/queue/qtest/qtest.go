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
	factory func(swarm.System, *swarm.Policy, chan string, string) swarm.Enqueue,
) {
	t.Helper()
	eff := make(chan string, 1)

	t.Run("Typed", func(t *testing.T) {
		cfg := swarm.DefaultPolicy()
		sys := system.NewSystem("test-system")
		enq := factory(sys, cfg, eff, Category)
		q := sys.Queue("test-queue", enq, nil, cfg)

		out, _ := queue.Enqueue[Note](q)
		if err := sys.Listen(); err != nil {
			panic(err)
		}

		out <- Note{Some: "message"}
		it.Ok(t).
			If(<-eff).Equal(Message)

		sys.Close()
	})

	t.Run("Bytes", func(t *testing.T) {
		cfg := swarm.DefaultPolicy()
		sys := system.NewSystem("test-system")
		enq := factory(sys, cfg, eff, Category)
		q := sys.Queue("test-queue", enq, nil, cfg)

		out, _ := bytes.Enqueue(q, Category)
		if err := sys.Listen(); err != nil {
			panic(err)
		}

		out <- []byte(Message)
		it.Ok(t).
			If(<-eff).Equal(Message)

		sys.Close()
	})

	t.Run("Events", func(t *testing.T) {
		cfg := swarm.DefaultPolicy()
		sys := system.NewSystem("test-system")
		enq := factory(sys, cfg, eff, Category)
		q := sys.Queue("test-queue", enq, nil, cfg)

		out, _ := events.Enqueue[*Note, swarm.Event[*Note]](q)
		if err := sys.Listen(); err != nil {
			panic(err)
		}

		out <- &swarm.Event[*Note]{
			Object: &Note{Some: "message"},
		}
		var val swarm.Event[*Note]
		err := json.Unmarshal([]byte(<-eff), &val)

		it.Ok(t).
			IfNil(err).
			If(*val.Object).Equal(Note{Some: "message"}).
			If(string(val.Type)).Equal("note:Event[*swarm.Note").
			IfTrue(val.ID != "").
			IfTrue(val.Created != "")

		sys.Close()
	})
}

func TestDequeue(
	t *testing.T,
	factory func(swarm.System, *swarm.Policy, chan string, string, string, string) swarm.Dequeue,
) {
	t.Helper()
	eff := make(chan string, 1)

	t.Run("Typed", func(t *testing.T) {
		note := Note{Some: "message"}
		message, _ := json.Marshal(note)

		cfg := swarm.DefaultPolicy()
		sys := system.NewSystem("test-system")
		deq := factory(sys, cfg, eff, Category, string(message), Receipt)
		q := sys.Queue("test-queue", nil, deq, cfg)

		msg, ack := queue.Dequeue[Note](q)
		if err := sys.Listen(); err != nil {
			panic(err)
		}

		val := <-msg
		ack <- val

		it.Ok(t).
			If(val.Object).Equal(note).
			If(<-eff).Equal(Receipt)

		sys.Close()
	})

	t.Run("Bytes", func(t *testing.T) {
		cfg := swarm.DefaultPolicy()
		sys := system.NewSystem("test-system")
		deq := factory(sys, cfg, eff, Category, Message, Receipt)
		q := sys.Queue("test-queue", nil, deq, cfg)

		msg, ack := bytes.Dequeue(q, Category)
		if err := sys.Listen(); err != nil {
			panic(err)
		}

		val := <-msg
		ack <- val

		it.Ok(t).
			If(string(val.Object)).Equal(Message).
			If(<-eff).Equal(Receipt)

		sys.Close()
	})

	t.Run("Events", func(t *testing.T) {
		event := swarm.Event[*Note]{
			ID:          "id",
			Type:        "type",
			Agent:       "agent",
			Participant: "user",
			Created:     "created",
			Object:      &Note{Some: "message"},
		}
		message, _ := json.Marshal(event)

		cfg := swarm.DefaultPolicy()
		sys := system.NewSystem("test-system")
		deq := factory(sys, cfg, eff, "note:Event[*swarm.Note]", string(message), Receipt)
		q := sys.Queue("test-queue", nil, deq, cfg)

		msg, ack := events.Dequeue[*Note, swarm.Event[*Note]](q)
		if err := sys.Listen(); err != nil {
			panic(err)
		}

		val := <-msg
		ack <- val

		it.Ok(t).
			If(val.Object).Equal(&Note{Some: "message"}).
			If(string(val.Agent)).Equal("agent").
			If(string(val.Participant)).Equal("user").
			If(string(val.Type)).Equal("type").
			If(string(val.ID)).Equal("id").
			If(string(val.Created)).Equal("created").
			If(<-eff).Equal(Receipt)

		sys.Close()
	})
}
