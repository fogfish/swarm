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

	t.Run("Success", func(t *testing.T) {
		out <- swarm.Bytes(Message)

		it.Ok(t).
			If(<-eff).Equal(Message)
	})

	t.Run("Failure", func(t *testing.T) {
		out, err := queue.Send("Some Other")
		out <- swarm.Bytes(Message)

		it.Ok(t).
			If(<-err).Equal(swarm.Bytes(Message))
	})

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

	msg, ack := queue.Recv(Category)
	sys.Listen()

	t.Run("Success", func(t *testing.T) {
		val := <-msg
		ack <- val

		it.Ok(t).
			If(val.Bytes()).Equal([]byte(Message)).
			If(<-eff).Equal(Receipt)
	})

	sys.Stop()
}
