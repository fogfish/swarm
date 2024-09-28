package enqueue_test

import (
	"context"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/enqueue"
	"github.com/fogfish/swarm/kernel"
)

func TestNewTypes(t *testing.T) {
	mock := mockEmitter(10)
	k := kernel.NewEnqueuer(mock, swarm.Config{})

	q := enqueue.NewTyped[User](k)
	q.Enq(context.Background(), User{ID: "id", Text: "user"})

	it.Then(t).Should(
		it.Equal(mock.val, `{"id":"id","text":"user"}`),
	)

	k.Close()
}

func TestNewBytes(t *testing.T) {
	mock := mockEmitter(10)
	k := kernel.NewEnqueuer(mock, swarm.Config{})

	q := enqueue.NewBytes(k, "User")
	q.Enq(context.Background(), []byte(`{"id":"id","text":"user"}`))

	it.Then(t).Should(
		it.Equal(mock.val, `{"id":"id","text":"user"}`),
	)

	k.Close()
}
