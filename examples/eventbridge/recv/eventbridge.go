package main

import (
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue"
	"github.com/fogfish/swarm/queue/eventbridge"
)

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func main() {
	sys := eventbridge.NewSystem("swarm-example-eventbridge")
	q := eventbridge.Must(eventbridge.New(sys, "swarm-test"))

	go actor("a").handle(queue.Recv[Note](q))
	// go actor("b").handle(queue.Recv("eventbridge.test.b"))
	// go actor("c").handle(queue.Recv("eventbridge.test.c"))

	if err := sys.Listen(); err != nil {
		panic(err)
	}

	sys.Wait()
}

//
type actor string

func (a actor) handle(rcv <-chan *swarm.MsgG[Note], ack chan<- *swarm.MsgG[Note]) {
	for msg := range rcv {
		logger.Debug("event on %s > %+v", a, msg.Object)
		ack <- msg
	}
}
