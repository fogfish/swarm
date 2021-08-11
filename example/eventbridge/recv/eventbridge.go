package main

import (
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue/eventbridge"
)

func main() {
	sys := swarm.New("test")
	queue := swarm.Must(eventbridge.New(sys, "test"))

	go (&actor{id: "a"}).onEvent(queue.Recv("eventbridge.test.a"))
	go (&actor{id: "b"}).onEvent(queue.Recv("eventbridge.test.b"))
	go (&actor{id: "c"}).onEvent(queue.Recv("eventbridge.test.c"))

	sys.Wait()
}

//
//
type actor struct {
	id string
}

func (a *actor) onEvent(rcv <-chan swarm.Msg, ack chan<- swarm.Msg) {
	for msg := range rcv {
		logger.Debug("event on %s > %s", a.id, msg.Bytes())
		ack <- msg
	}
}
