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
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/backoff"
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/bytes"
	"github.com/fogfish/swarm/queue/events"
)

const (
	Category = "Note"
	Message  = "{\"some\":\"message\"}"
	Event    = ""
	Receipt  = "0x123456789abcdef"
)

var (
	retry200ms = swarm.WithRetry(backoff.Exp(10*time.Millisecond, 4, 0.2))
)

type Note struct {
	Some string `json:"some"`
}

type User struct {
	Some string `json:"some"`
}

type effect = chan string
type queueName = string
type category = string
type message = string
type receipt = string
type enqueue = func(effect, queueName, category, ...swarm.Option) swarm.Broker
type dequeue = func(effect, queueName, category, message, receipt, ...swarm.Option) swarm.Broker

//
//
func TestEnqueueTyped(t *testing.T, factory enqueue) {
	t.Helper()
	eff := make(chan string, 1)

	q := factory(eff, "test-queue", Category, retry200ms)

	note, _ := queue.Enqueue[Note](q)
	user, dlq := queue.Enqueue[User](q)

	t.Run("Enqueue", func(t *testing.T) {
		note <- Note{Some: "message"}

		select {
		case in := <-eff:
			it.Then(t).Should(it.Equal(in, Message))
		case <-time.After(50 * time.Millisecond):
			t.Error("failed to send message")
		}
	})

	t.Run("EnqueueFailed", func(t *testing.T) {
		user <- User{Some: "message"}

		select {
		case in := <-dlq:
			it.Then(t).Should(it.Equal(in, User{Some: "message"}))
		case <-time.After(200 * time.Millisecond):
			t.Error("failed to receive dlq message message")
		}
	})

	q.Close()
}

//
//
func TestEnqueueBytes(t *testing.T, factory enqueue) {
	t.Helper()
	eff := make(chan string, 1)

	q := factory(eff, "test-queue", Category, retry200ms)

	note, _ := bytes.Enqueue(q, Category)
	user, dlq := bytes.Enqueue(q, "User")

	t.Run("Enqueue", func(t *testing.T) {
		note <- []byte(Message)

		select {
		case in := <-eff:
			it.Then(t).Should(it.Equal(in, Message))
		case <-time.After(50 * time.Millisecond):
			t.Error("failed to send message")
		}
	})

	t.Run("EnqueueFailed", func(t *testing.T) {
		user <- []byte(Message)

		select {
		case in := <-dlq:
			it.Then(t).Should(it.Equiv(in, []byte(Message)))
		case <-time.After(200 * time.Millisecond):
			t.Error("failed to receive dlq message message")
		}
	})

	q.Close()
}

//
//
func TestEnqueueEvent(t *testing.T, factory enqueue) {
	t.Helper()
	eff := make(chan string, 1)

	q := factory(eff, "test-queue", "note:Event[*swarm.Note]", retry200ms)

	note, _ := events.Enqueue[*Note, swarm.Event[*Note]](q)
	user, dlq := events.Enqueue[*User, swarm.Event[*User]](q)

	t.Run("Enqueue", func(t *testing.T) {
		note <- &swarm.Event[*Note]{
			Object: &Note{Some: "message"},
		}

		select {
		case raw := <-eff:
			var val swarm.Event[*Note]
			err := json.Unmarshal([]byte(raw), &val)

			it.Then(t).
				Should(it.Nil(err)).
				Should(it.Equal(*val.Object, Note{Some: "message"})).
				Should(it.Equal(val.Type, "note:Event[*swarm.Note")). // TODO: fogfish/curie#29
				ShouldNot(it.Equal(len(val.ID), 0)).
				ShouldNot(it.Equal(len(val.Created), 0))

		case <-time.After(50 * time.Millisecond):
			t.Error("failed to send message")
		}
	})

	t.Run("EnqueueFailed", func(t *testing.T) {
		user <- &swarm.Event[*User]{
			Object: &User{Some: "message"},
		}

		select {
		case in := <-dlq:
			it.Then(t).Should(it.Equal(*in.Object, User{Some: "message"}))
		case <-time.After(200 * time.Millisecond):
			t.Error("failed to receive dlq message message")
		}
	})

	q.Close()
}

//
//
func TestDequeueTyped(t *testing.T, factory dequeue) {
	t.Helper()
	eff := make(chan string, 1)

	t.Run("Typed", func(t *testing.T) {
		q := factory(eff, "test-queue", Category, Message, Receipt, retry200ms)

		msg, ack := queue.Dequeue[Note](q)
		go q.Await()

		val := <-msg
		ack <- val

		it.Then(t).
			Should(it.Equal(val.Object, Note{Some: "message"})).
			Should(it.Equal(<-eff, Receipt))

		q.Close()
	})
}

//
//
func TestDequeueBytes(t *testing.T, factory dequeue) {
	t.Helper()
	eff := make(chan string, 1)

	t.Run("Typed", func(t *testing.T) {
		q := factory(eff, "test-queue", Category, Message, Receipt, retry200ms)

		msg, ack := bytes.Dequeue(q, Category)
		go q.Await()

		val := <-msg
		ack <- val

		it.Then(t).
			Should(it.Equiv(val.Object, []byte(Message))).
			Should(it.Equal(<-eff, Receipt))

		q.Close()
	})
}

//
//
func TestDequeueEvent(t *testing.T, factory dequeue) {
	t.Helper()
	eff := make(chan string, 1)

	t.Run("Typed", func(t *testing.T) {
		event := swarm.Event[*Note]{
			ID:          "id",
			Type:        "type",
			Agent:       "agent",
			Participant: "user",
			Created:     "created",
			Object:      &Note{Some: "message"},
		}
		message, _ := json.Marshal(event)

		q := factory(eff, "test-queue", "note:Event[*swarm.Note]", string(message), Receipt, retry200ms)

		msg, ack := events.Dequeue[*Note, swarm.Event[*Note]](q)
		go q.Await()

		val := <-msg
		ack <- val

		it.Then(t).
			Should(it.Equal(*val.Object, Note{Some: "message"})).
			Should(it.Equal(val.Agent, "agent")).
			Should(it.Equal(val.Participant, "user")).
			Should(it.Equal(val.Type, "type")).
			Should(it.Equal(val.ID, "id")).
			Should(it.Equal(val.Created, "created")).
			Should(it.Equal(<-eff, Receipt))

		q.Close()
	})
}
