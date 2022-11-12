//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

type Broker interface {
	Config() *Config
	Close()
	DSync()
	Await()
	Enqueue(string, Channel) Enqueue
	Dequeue(string, Channel) Dequeue
}

type Enqueue interface {
	Enq(Bag) error
}

type Dequeue interface {
	Deq(string) (Bag, error)
	Ack(Bag) error
}
