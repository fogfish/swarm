package qtest

import (
	"testing"

	"github.com/fogfish/it"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/system"
	"github.com/fogfish/swarm/queue"
)

const (
	Category = "Note"
	Message  = "{\"some\":\"message\"}"
	Receipt  = "0x123456789abcdef"
)

type Note struct {
	Some string `json:"some"`
}

func TestSend(
	t *testing.T,
	factory func(swarm.System, *swarm.Policy, chan string) (swarm.Sender, swarm.Recver),
) {
	t.Helper()

	eff := make(chan string, 1)
	sys := system.NewSystem("qtest")
	q := sys.Queue(factory(sys, swarm.DefaultPolicy(), eff))

	out, _ := queue.Send[Note](q)
	if err := sys.Listen(); err != nil {
		panic(err)
	}

	t.Run("Success", func(t *testing.T) {
		out <- Note{Some: "message"}
		it.Ok(t).
			If(<-eff).Equal(Message)
	})

	// t.Run("Failure", func(t *testing.T) {
	// 	out, err := queue.Send("Some Other")
	// 	out <- swarm.Bytes(Message)

	// 	it.Ok(t).
	// 		If(<-err).Equal(swarm.Bytes(Message))
	// })

	sys.Stop()
}

func TestRecv(
	t *testing.T,
	factory func(swarm.System, *swarm.Policy, chan string) (swarm.Sender, swarm.Recver),
) {
	t.Helper()

	eff := make(chan string, 1)
	sys := system.NewSystem("qtest")
	q := sys.Queue(factory(sys, swarm.DefaultPolicy(), eff))

	msg, ack := queue.Recv[Note](q)
	if err := sys.Listen(); err != nil {
		panic(err)
	}

	t.Run("Success", func(t *testing.T) {
		val := <-msg
		ack <- val

		it.Ok(t).
			If(val.Object).Equal(Note{Some: "message"}).
			If(<-eff).Equal(Receipt)
	})

	sys.Stop()
}
