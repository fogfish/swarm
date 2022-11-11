package events

import (
	"encoding/json"
	"strings"
	"time"
	"unsafe"

	"github.com/fogfish/guid"
	"github.com/fogfish/swarm"
)

type Queue[T any, E swarm.EventKind[T]] interface {
	Enqueue(*E) error
}

//
type queue[T any, E swarm.EventKind[T]] struct {
	cat  string
	conf *swarm.Config
	sock swarm.Enqueue

	offID, offType, offCreated uintptr
}

func (q queue[T, E]) Sync()  {}
func (q queue[T, E]) Close() {}

func (q queue[T, E]) Enqueue(object *E) error {
	evt := unsafe.Pointer(object)

	// patch ID
	if len(*(*string)(unsafe.Pointer(uintptr(evt) + q.offID))) == 0 {
		id := guid.G.K(guid.Clock).String()
		*(*string)(unsafe.Pointer(uintptr(evt) + q.offID)) = id
	}

	// patch Type
	if len(*(*string)(unsafe.Pointer(uintptr(evt) + q.offType))) == 0 {
		*(*string)(unsafe.Pointer(uintptr(evt) + q.offType)) = q.cat
	}

	// patch Created
	if len(*(*string)(unsafe.Pointer(uintptr(evt) + q.offCreated))) == 0 {
		t := time.Now().Format(time.RFC3339)
		*(*string)(unsafe.Pointer(uintptr(evt) + q.offCreated)) = t
	}

	msg, err := json.Marshal(object)
	if err != nil {
		return err
	}

	bag := swarm.Bag{Category: q.cat, Object: msg}
	err = q.conf.Backoff.Retry(func() error { return q.sock.Enq(bag) })
	if err != nil {
		return err
	}

	return nil
}

//
func New[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) Queue[T, E] {
	cat := strings.ToLower(typeOf[T]()) + ":" + typeOf[E]()
	if len(category) > 0 {
		cat = category[0]
	}

	//
	// building memory layout to make unsafe struct reading
	kindT := swarm.Event[T]{}
	offID, offType, offCreated :=
		unsafe.Offsetof(kindT.ID),
		unsafe.Offsetof(kindT.Type),
		unsafe.Offsetof(kindT.Created)

	queue := &queue[T, E]{
		cat:        cat,
		conf:       q.Config(),
		offID:      offID,
		offType:    offType,
		offCreated: offCreated,
	}
	sock, err := q.Enqueue(cat, queue)
	if err != nil {
		panic(err)
	}
	queue.sock = sock

	return queue
}
