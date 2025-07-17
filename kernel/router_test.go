//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import (
	"context"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
)

func TestMsgRoute(t *testing.T) {
	r := newMsgRouter(
		make(chan swarm.Msg[string], 1),
		encoding.ForTyped[string](),
	)

	r.Route(context.Background(), swarm.Bag{Object: []byte(`"1"`)})
	it.Then(t).Should(
		it.Equal((<-r.ch).Object, `1`),
	)
}

func TestEvtRoute(t *testing.T) {
	type E = swarm.Event[swarm.Meta, string]

	r := newEvtRouter(
		make(chan E, 1),
		encoding.ForEvent[E]("realm", "agent"),
	)

	r.Route(context.Background(), swarm.Bag{Object: []byte(`{"data": "1"}`)})
	it.Then(t).Should(
		it.Equal(*(<-r.ch).Data, `1`),
	)
}
