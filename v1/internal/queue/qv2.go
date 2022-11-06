package queue

import "github.com/fogfish/swarm"

type Enqueue interface {
	End(swarm.Bag) error
}

type Dequeue interface {
	Deq() (swarm.Bag, error)
	Ack(swarm.Bag)
}
