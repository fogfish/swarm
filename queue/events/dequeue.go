package events

import (
	"encoding/json"
	"strings"
	"unsafe"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Dequeue ...
*/
func Dequeue[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) (<-chan *E, chan<- *E) {
	conf := q.Config()
	ch := swarm.NewEvtDeqCh[T, E](conf.DequeueCapacity)

	cat := strings.ToLower(typeOf[T]()) + ":" + typeOf[E]()
	if len(category) > 0 {
		cat = category[0]
	}

	//
	// building memory layout to make unsafe struct reading
	kindT := swarm.Event[T]{}
	offDigest := unsafe.Offsetof(kindT.Digest)

	sock, err := q.Dequeue(cat, ch)
	if err != nil {
		panic(err)
	}

	pipe.ForEach(ch.Ack, func(object *E) {
		evt := unsafe.Pointer(object)
		digest := *(*string)(unsafe.Pointer(uintptr(evt) + offDigest))

		// TODO: Error logging
		conf.Backoff.Retry(func() error {
			return sock.Ack(swarm.Bag{
				Category: cat,
				Digest:   digest,
			})
		})
	})

	pipe.Emit(ch.Msg, q.Config().PollFrequency, func() (*E, error) {
		var bag swarm.Bag
		err := conf.Backoff.Retry(func() (err error) {
			bag, err = sock.Deq(cat)
			return
		})
		if err != nil {
			return nil, err
		}

		evt := new(E)
		if err := json.Unmarshal(bag.Object, evt); err != nil {
			return nil, err
		}

		ptr := unsafe.Pointer(evt)
		*(*string)(unsafe.Pointer(uintptr(ptr) + offDigest)) = bag.Digest

		return evt, nil
	})

	return ch.Msg, ch.Ack
}
