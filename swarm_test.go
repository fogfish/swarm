package swarm_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue/sqs"
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

/*
	TODO:

	sys := swarm.New("sys")
	q, _ := ephemeral.New(sys)

	q.Recv("cat")
	q.Send("cat")

	TODO:
		- register queue at system

*/

func TestX(t *testing.T) {
	sys := swarm.New("sys")
	// q, _ := ephemeral.New(sys)
	a, _ := sqs.New(sys, "test")

	go actor(a.Recv("catx"))

	// go actor(q.Recv("cat"))

	// send := q.Send("cat")
	// send <- []byte("axx")
	// send <- []byte("bxx")
	ax, ar := a.Send("catx")
	go erractor(ar)

	ax <- swarm.Bytes("cxx")

	time.Sleep(60 * time.Minute)
	sys.Stop()

	// time.Sleep(5 * time.Second)
}

func actor(mbox <-chan swarm.Msg, ack chan<- swarm.Msg) {
	for x := range mbox {
		fmt.Printf("%+v\n", string(x.Bytes()))
		ack <- x
	}
}

func erractor(mbox <-chan swarm.Msg) {
	for x := range mbox {
		fmt.Printf("error >>> %+v\n", string(x.Bytes()))
	}
}

/*
type M interface {
	Bytes() []byte
}

type X []byte

func (x X) Bytes() []byte {
	return x
}



func x(a M) M {
	return a
}

func TestX(t *testing.T) {
	v := x(X([]byte("xxxxx")))

	fmt.Println(bytes.NewBuffer(v.Bytes()))
}
*/
