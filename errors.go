package swarm

import "github.com/fogfish/faults"

const (
	ErrServiceIO = faults.Type("service i/o failed")
	ErrEnqueue   = faults.Type("enqueue is failed")
	ErrDequeue   = faults.Type("dequeue is failed")
)
