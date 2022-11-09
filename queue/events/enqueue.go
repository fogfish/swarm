package events

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/fogfish/guid"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/pipe"
)

/*

Enqueue creates pair of channels to send messages and dead-letter queue
*/
func Enqueue[T any, E swarm.EventKind[T]](q swarm.Broker, category ...string) (chan<- *E, <-chan *E) {
	ch := swarm.NewEvtEnqCh[T, E](q.Config().EnqueueCapacity)

	cat := strings.ToLower(typeOf[T]()) + ":" + typeOf[E]()

	//
	// building memory layout to make unsafe struct reading
	kindT := swarm.Event[T]{}
	offID, offType, offCreated :=
		unsafe.Offsetof(kindT.ID),
		unsafe.Offsetof(kindT.Type),
		unsafe.Offsetof(kindT.Created)

	sock, err := q.Enqueue(cat, ch)
	if err != nil {
		panic(err)
	}

	pipe.ForEach(ch.Msg, func(object *E) {
		evt := unsafe.Pointer(object)

		// patch ID
		if len(*(*string)(unsafe.Pointer(uintptr(evt) + offID))) == 0 {
			id := guid.G.K(guid.Clock).String()
			*(*string)(unsafe.Pointer(uintptr(evt) + offID)) = id
		}

		// patch Type
		if len(*(*string)(unsafe.Pointer(uintptr(evt) + offType))) == 0 {
			*(*string)(unsafe.Pointer(uintptr(evt) + offType)) = cat
		}

		// patch Created
		if len(*(*string)(unsafe.Pointer(uintptr(evt) + offCreated))) == 0 {
			t := time.Now().Format(time.RFC3339)
			*(*string)(unsafe.Pointer(uintptr(evt) + offCreated)) = t
		}

		msg, err := json.Marshal(object)
		if err != nil {
			ch.Err <- object
			return
		}

		err = sock.Enq(swarm.Bag{
			Category: cat,
			Object:   msg,
		})
		if err != nil {
			ch.Err <- object
		}
	})

	return ch.Msg, ch.Err
}

func typeOf[T any]() string {
	//
	// TODO: fix
	//   Action[*swarm.User] if container type is used
	//

	typ := reflect.TypeOf(*new(T))
	cat := typ.Name()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().Name()
	}

	return cat
}
