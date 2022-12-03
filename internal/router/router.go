//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package router

import (
	"fmt"
	"sync"
	"time"

	"github.com/fogfish/swarm"
)

type Router struct {
	sync.Mutex
	config *swarm.Config
	sack   chan swarm.Bag
	sock   map[string]chan swarm.Bag
	acks   map[string]struct{}
	onAck  func(swarm.Bag) error
}

func New(config *swarm.Config, onAck func(swarm.Bag) error) *Router {
	return &Router{
		config: config,
		sack:   make(chan swarm.Bag, config.DequeueCapacity),
		sock:   make(map[string]chan swarm.Bag),
		acks:   map[string]struct{}{},
		onAck:  onAck,
	}
}

func (router *Router) Register(category string) {
	router.Lock()
	defer router.Unlock()

	router.sock[category] = make(chan swarm.Bag, router.config.DequeueCapacity)
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
				err := router.config.Backoff.Retry(func() error {
					return router.onAck(bag)
				})
				if err != nil {
					return err
				}
			}

			delete(router.acks, bag.Digest)
			if len(router.acks) == 0 {
				if router.config.HookCommit != nil {
					router.config.HookCommit()
				}
				return nil
			}
		case <-time.After(d):
			router.acks = map[string]struct{}{}
			return fmt.Errorf("timeout message ack")
		}
	}
}
