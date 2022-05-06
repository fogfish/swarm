package queue

import (
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/queue/eventbridge"
	"github.com/fogfish/swarm/internal/queue/sqs"
	"github.com/fogfish/swarm/internal/system"
)

/*

New creates new queueing system
*/
func System(id string) swarm.System {
	return system.NewSystem(id)
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

func policy(opts []*swarm.Policy) *swarm.Policy {
	if len(opts) > 0 {
		return opts[0]
	}
	return swarm.DefaultPolicy()
}

/*

SQS ...
*/
func SQS(sys swarm.System, queue string, opts ...*swarm.Policy) (swarm.Queue, error) {
	q, err := sqs.New(sys, queue, policy(opts))
	if err != nil {
		return nil, err
	}

	return sys.Queue(q), nil
}

/*

EventBridge ...
*/
func EventBridge(sys swarm.System, queue string, opts ...*swarm.Policy) (swarm.Queue, error) {
	q, err := eventbridge.New(sys, queue, policy(opts))
	if err != nil {
		return nil, err
	}

	return sys.Queue(q), nil
}
