//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/queue/eventbridge"
	"github.com/fogfish/swarm/internal/system"
)

/*

NewSystem creates new queueing system
*/
func NewSystem(id string) swarm.System {
	return system.NewSystem(id)
}

/*

New create a logical queue (topic) over eventbridge
*/
func New(sys swarm.System, queue string, opts ...*swarm.Policy) (swarm.Queue, error) {
	policy := swarm.DefaultPolicy()
	if len(opts) > 0 {
		policy = opts[0]
	}

	awscli, err := system.NewSession()
	if err != nil {
		return nil, err
	}

	enq := eventbridge.NewEnqueue(sys, queue, policy, awscli)
	deq := eventbridge.NewDequeue(sys, queue, policy)

	return sys.Queue(queue, enq, deq, policy), nil
}

/*

Must ensures successful creation of queue
*/
func Must(q swarm.Queue, err error) swarm.Queue {
	if err != nil {
		panic(err)
	}

	return q
}
