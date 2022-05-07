package main

import (
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue/eventbridge"
)

func main() {
	sys := eventbridge.NewSystem("swarm-example-eventbridge")
	queue := eventbridge.Must(eventbridge.New(sys, "swarm-test"))

	go actor("a").handle(queue.Recv("eventbridge.test.a"))
	go actor("b").handle(queue.Recv("eventbridge.test.b"))
	go actor("c").handle(queue.Recv("eventbridge.test.c"))

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	sys.Wait()
}

//
type actor string

func (a actor) handle(rcv <-chan swarm.Object, ack chan<- swarm.Object) {
	for msg := range rcv {
		logger.Debug("event on %s > %s", a, msg.Bytes())
		ack <- msg
	}
}
