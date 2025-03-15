//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package dequeue_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/dequeue"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

// controls yield time before kernel is closed
const yield_before_close = 5 * time.Millisecond

type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func TestDequeueType(t *testing.T) {
	user := User{ID: "id", Text: "user"}

	k := kernel.NewDequeuer(mockCathode(user), swarm.Config{})
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

func TestDequeueBytes(t *testing.T) {
	user := User{ID: "id", Text: "user"}

	k := kernel.NewDequeuer(mockCathode(user), swarm.Config{})
	go func() {
		time.Sleep(yield_before_close)
		k.Close()
	}()

	var msg swarm.Msg[[]byte]
	rcv, ack := dequeue.Bytes(k, encoding.NewCodecByte("User"))

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

type cathode struct {
	cat  string
	user User
}

func mockCathode(user User) cathode {
	return cathode{
		cat:  swarm.TypeOf[User](),
		user: user,
	}
}

func (c cathode) Ack(ctx context.Context, digest string) error {
	return nil
}

func (c cathode) Err(ctx context.Context, digest string, err error) error {
	return nil
}

func (c cathode) Ask(context.Context) ([]swarm.Bag, error) {
	data, err := json.Marshal(c.user)
	if err != nil {
		return nil, err
	}

	bag := []swarm.Bag{{Category: c.cat, Digest: "1", Object: data}}
	return bag, nil
}
