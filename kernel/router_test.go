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

func TestRoute(t *testing.T) {
	r := router[string]{
		ch:    make(chan swarm.Msg[string], 1),
		codec: encoding.ForTyped[string](),
	}

	r.Route(context.Background(), swarm.Bag{Object: []byte(`"1"`)})
	it.Then(t).Should(
		it.Equal((<-r.ch).Object, `1`),
	)
}
