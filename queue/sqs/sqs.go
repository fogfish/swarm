package sqs

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/queue/sqs"
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

	enq := sqs.NewEnqueue(sys, queue, policy, awscli)
	deq := sqs.NewDequeue(sys, queue, policy, awscli)

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
