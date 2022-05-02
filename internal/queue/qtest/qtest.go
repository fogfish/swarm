package qtest

import (
	"testing"

	"github.com/fogfish/it"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue"
)

const (
	Category = "q.test"
	Message  = "{\"some\":\"message\"}"
	Receipt  = "0x123456789abcdef"
)

func TestSend(
	t *testing.T,
	factory func(swarm.System, *swarm.Policy, chan string) swarm.EventBus,
) {
	t.Helper()

	eff := make(chan string, 1)
	sys := queue.System("qtest")
	queue := sys.Queue(factory(sys, swarm.DefaultPolicy(), eff))

	out, _ := queue.Send(Category)

	sys.Listen()

	out <- swarm.Bytes(Message)
	it.Ok(t).
		If(<-eff).Equal(Message)

	sys.Stop()
}

func TestRecv(
	t *testing.T,
	factory func(swarm.System, *swarm.Policy, chan string) swarm.EventBus,
) {
	t.Helper()

	eff := make(chan string, 1)
	sys := queue.System("qtest")
	queue := sys.Queue(factory(sys, swarm.DefaultPolicy(), eff))

	msg, _ := queue.Recv(Category)

	sys.Listen()

	val := <-msg
	it.Ok(t).
		If(val.Bytes()).Equal([]byte(Message))

	sys.Stop()
}
