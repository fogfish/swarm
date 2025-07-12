//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package listen_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
	dequeue "github.com/fogfish/swarm/listen"
)

// controls yield time before kernel is closed
const yield_before_close = 5 * time.Millisecond

type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Evt = swarm.Event[swarm.Meta, User]

func TestDequeueType(t *testing.T) {
	cfg := swarm.NewConfig()
	cfg.PollFrequency = 1 * time.Millisecond

	user := User{ID: "id", Text: "user"}

	k := kernel.NewListener(mockCathode("User", user), cfg)
	go func() {
		time.Sleep(yield_before_close)
		k.Close()
	}()

	var msg swarm.Msg[User]
	rcv, ack := dequeue.Typed[User](k)

	go func() {
		msg = <-rcv
		ack <- msg
	}()
	k.Await()

	it.Then(t).Should(
		it.Equal(msg.Category, "User"),
		it.Equal(msg.Digest, "1"),
		it.Equal(msg.Object.ID, "id"),
		it.Equal(msg.Object.Text, "user"),
	)
}

func TestDequeueEvent(t *testing.T) {
	cfg := swarm.NewConfig()
	cfg.PollFrequency = 1 * time.Millisecond

	obj := Evt{
		Meta: &swarm.Meta{Type: "User"},
		Data: &User{ID: "id", Text: "user"},
	}

	k := kernel.NewListener(mockCathode("User", obj), cfg)
	go func() {
		time.Sleep(yield_before_close)
		k.Close()
	}()

	var evt swarm.Msg[Evt]
	rcv, ack := dequeue.Event[Evt](k)

	go func() {
		evt = <-rcv
		ack <- evt
	}()
	k.Await()

	it.Then(t).Should(
		it.Equal(evt.Category, "User"),
		it.Equal(evt.Digest, "1"),
		it.Equal(evt.Object.Meta.Type, "User"),
		it.Equal(evt.Object.Data.ID, "id"),
		it.Equal(evt.Object.Data.Text, "user"),
	)
}

func TestDequeueBytes(t *testing.T) {
	cfg := swarm.NewConfig()
	cfg.PollFrequency = 1 * time.Millisecond

	user := User{ID: "id", Text: "user"}

	k := kernel.NewListener(mockCathode("User", user), cfg)
	go func() {
		time.Sleep(yield_before_close)
		k.Close()
	}()

	var msg swarm.Msg[[]byte]
	rcv, ack := dequeue.Bytes(k, encoding.ForBytes("User"))

	go func() {
		msg = <-rcv
		ack <- msg
	}()
	k.Await()

	it.Then(t).Should(
		it.Equal(msg.Category, "User"),
		it.Equal(msg.Digest, "1"),
		it.Equal(string(msg.Object), `{"id":"id","text":"user"}`),
	)
}

//------------------------------------------------------------------------------

type cathode[T any] struct {
	cat string
	obj T
}

func mockCathode[T any](cat string, obj T) cathode[T] {
	return cathode[T]{
		cat: cat,
		obj: obj,
	}
}

func (c cathode[T]) Ack(ctx context.Context, digest string) error {
	return nil
}

func (c cathode[T]) Err(ctx context.Context, digest string, err error) error {
	return nil
}

func (c cathode[T]) Ask(context.Context) ([]swarm.Bag, error) {
	data, err := json.Marshal(c.obj)
	if err != nil {
		return nil, err
	}

	bag := []swarm.Bag{{Category: c.cat, Digest: "1", Object: data}}
	return bag, nil
}
