package swarm_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue/ephemeral"
)

/*
func TestOneChannelMultipleReader(t *testing.T) {
	ch := make(chan int)

	go func() {
		// for x := range ch {
		// 	fmt.Printf("a %v\n", x)
		// }
		for {
			select {
			case x := <-ch:
				fmt.Printf("a %v\n", x)
			}
		}
	}()

	go func() {
		// for x := range ch {
		// 	fmt.Printf("b %v\n", x)
		// }
		for {
			select {
			case x := <-ch:
				fmt.Printf("b %v\n", x)
			}
		}

	}()

	ch <- 1
	ch <- 2
	ch <- 3

	time.Sleep(5 * time.Second)
}
*/

// func TestInMem(t *testing.T) {
// 	recv, send := swarm.InMem()

// 	send <- &swarm.Message{Object: []byte("xxx")}
// 	v := <-recv

// 	fmt.Printf("%+v\n", v)
// }

func TestX(t *testing.T) {
	sys := swarm.New("sys")
	q, _ := ephemeral.New(sys)

	/*
		TODO:

		sys := swarm.New("sys")
		q, _ := ephemeral.New(sys)

		q.Recv("cat")
		q.Send("cat")

		TODO:
			- register queue at system

	*/

	go actor(q.Recv("cat"))

	// sys.Listen("xxx")

	send := q.Send("cat")
	send <- []byte("axx")
	send <- []byte("bxx")
	q.Send("catx") <- []byte("cxx")

	time.Sleep(1 * time.Second)
	sys.Stop()

	// &swarm.Message{Category: "cat", Object: }
	// send <- &swarm.Message{Category: "cat", Object: []byte("bxx")}
	// send <- &swarm.Message{Category: "cat", Object: []byte("cxx")}

	time.Sleep(5 * time.Second)
}

func actor(mbox <-chan []byte) {
	for x := range mbox {
		fmt.Printf("%+v\n", x)
	}
}
