//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package system

import (
	"context"
	"sync"

	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

type system struct {
	sync.Mutex
	queue map[string]*Queue

	id      string
	context context.Context
	cancel  context.CancelFunc
	logger  logger.Logger
}

var (
	_ swarm.System = (*system)(nil)
)

/*

New creates new queueing system
*/
func NewSystem(id string) swarm.System {
	sys := &system{
		id:      id,
		queue:   make(map[string]*Queue),
		context: context.Background(),
		logger:  logger.With(logger.Note{"system": id}),
	}

	sys.context, sys.cancel = context.WithCancel(sys.context)
	return sys
}

/*

Queue ...
*/
func (sys *system) Queue(id string, enq swarm.Enqueue, deq swarm.Dequeue, policy *swarm.Policy) swarm.Queue {
	sys.Lock()
	defer sys.Unlock()

	queue := NewQueue(sys, id, enq, deq, policy)
	sys.queue[id] = queue
	return queue
}

/*

ID return unique system ID
*/
func (sys *system) ID() string {
	return sys.id
}

/*

Listen ...
*/
func (sys *system) Listen() error {
	sys.logger.Info("system listening")

	for _, q := range sys.queue {
		if err := q.Listen(); err != nil {
			return err
		}
	}

	return nil
}

/*

Stop ...
*/
func (sys *system) Close() {
	for _, q := range sys.queue {
		q.Close()
	}

	sys.logger.Info("system closed")
	sys.cancel()
}

/*

Wait ...
*/
func (sys *system) Wait() {
	<-sys.context.Done()
}
