package events3

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

func TestReader(t *testing.T) {
	var bag []swarm.Bag
	bridge := &bridge{kernel.NewBridge(100 * time.Millisecond)}

	t.Run("New", func(t *testing.T) {
		q, err := NewReader()
		it.Then(t).Should(it.Nil(err))
		q.Close()
	})

	t.Run("Dequeue", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
			for _, m := range bag {
				bridge.Ack(context.Background(), m.Digest)
			}
		}()

		err := bridge.run(
			S3Event{
				Records: []json.RawMessage{[]byte(`{"sut":"test"}`)},
			},
		)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(len(bag), 1),
			it.Equal(bag[0].Category, Category),
			it.Equiv(bag[0].Object, []byte(`{"sut":"test"}`)),
		)
	})

	t.Run("Dequeue.Timeout", func(t *testing.T) {
		go func() {
			bag, _ = bridge.Ask(context.Background())
		}()

		err := bridge.run(
			S3Event{
				Records: []json.RawMessage{[]byte(`{"sut":"test"}`)},
			},
		)

		it.Then(t).ShouldNot(
			it.Nil(err),
		)
	})
}
