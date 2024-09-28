package enqueue_test

import (
	"context"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/enqueue"
	"github.com/fogfish/swarm/kernel"
)

// controls yield time before kernel is closed
const yield_before_close = 5 * time.Millisecond

type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func TestType(t *testing.T) {
	mock := mockEmitter(10)
	k := kernel.NewEnqueuer(mock, swarm.Config{})
	go func() {
		time.Sleep(yield_before_close)
		k.Close()
	}()

	snd, _ := enqueue.Typed[User](k)
	snd <- User{ID: "id", Text: "user"}

	k.Await()

	it.Then(t).Should(
		it.Equal(mock.val, `{"id":"id","text":"user"}`),
	)
}

func TestBytes(t *testing.T) {
	mock := mockEmitter(10)
	k := kernel.NewEnqueuer(mock, swarm.Config{})
	go func() {
		time.Sleep(yield_before_close)
		k.Close()
	}()

	snd, _ := enqueue.Bytes(k, "User")
	snd <- []byte(`{"id":"id","text":"user"}`)

	k.Await()

	it.Then(t).Should(
		it.Equal(mock.val, `{"id":"id","text":"user"}`),
	)

}

//------------------------------------------------------------------------------

type emitter struct {
	val string
}

func mockEmitter(wait int) *emitter {
	return &emitter{}
}

func (e *emitter) Enq(ctx context.Context, bag swarm.Bag) error {
	e.val = string(bag.Object)
	return nil
}
