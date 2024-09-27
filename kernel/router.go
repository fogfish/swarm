//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import "github.com/fogfish/swarm"

// Router is typed pair of message channel and codec
type router[T any] struct {
	ch    chan swarm.Msg[T]
	codec Decoder[T]
}

func (a router[T]) Route(bag swarm.Bag) error {
	obj, err := a.codec.Decode(bag.Object)
	if err != nil {
		return err
	}

	msg := swarm.Msg[T]{Ctx: bag.Ctx, Object: obj}
	a.ch <- msg
	return nil
}
