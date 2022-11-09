package router

import (
	"fmt"
	"sync"
	"time"

	"github.com/fogfish/swarm"
)

type Router struct {
	sync.Mutex
	sack  chan swarm.Bag
	sock  map[string]chan swarm.Bag
	acks  map[string]struct{}
	onAck func(swarm.Bag) error
}

func New(onAck func(swarm.Bag) error) *Router {
	return &Router{
		sack:  make(chan swarm.Bag),
		sock:  make(map[string]chan swarm.Bag),
		acks:  map[string]struct{}{},
		onAck: onAck,
	}
}

func (router *Router) Register(category string) {
	router.Lock()
	defer router.Unlock()

	// TODO: capacity
	router.sock[category] = make(chan swarm.Bag)
}

func (router *Router) Ack(bag swarm.Bag) error {
	router.sack <- bag
	return nil
}

func (router *Router) Deq(category string) (swarm.Bag, error) {
	bag := <-router.sock[category]
	return bag, nil
}

func (router *Router) Dispatch(bag swarm.Bag) error {
	sock, exists := router.sock[bag.Category]
	if !exists {
		return fmt.Errorf("not found category %s", bag.Category)
	}

	router.acks[bag.Digest] = struct{}{}
	sock <- bag

	return nil
}

func (router *Router) Await(d time.Duration) error {
	for {
		select {
		case bag := <-router.sack:
			if router.onAck != nil {
				if err := router.onAck(bag); err != nil {
					return err
				}
			}

			delete(router.acks, bag.Digest)
			if len(router.acks) == 0 {
				return nil
			}
		case <-time.After(d):
			router.acks = map[string]struct{}{}
			return fmt.Errorf("timeout message ack")
		}
	}
}
