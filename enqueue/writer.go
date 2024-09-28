package enqueue

import (
	"context"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

// Synchronous emitter of typed messages to the broker.
// It blocks the routine until the message is accepted by the broker.
type EmitterTyped[T any] struct {
	cat    string
	codec  kernel.Encoder[T]
	kernel *kernel.Enqueuer
}

// Creates synchronous typed emitter
func NewTyped[T any](q *kernel.Enqueuer, category ...string) *EmitterTyped[T] {
	return &EmitterTyped[T]{
		cat:    swarm.TypeOf[T](category...),
		codec:  encoding.NewCodecJson[T](),
		kernel: q,
	}
}

// Synchronously enqueue message to broker.
// It guarantees message to be send after return
func (q *EmitterTyped[T]) Enq(ctx context.Context, object T, cat ...string) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	category := q.cat
	if len(cat) > 0 {
		category = cat[0]
	}

	bag := swarm.Bag{
		Ctx:    swarm.NewContext(context.Background(), category, ""),
		Object: msg,
	}

	err = q.kernel.Emitter.Enq(ctx, bag)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

// Synchronous emitter of events to the broker.
// It blocks the routine until the event is accepted by the broker.
type EmitterEvent[T any, E swarm.EventKind[T]] struct {
	cat    string
	codec  kernel.Encoder[*E]
	kernel *kernel.Enqueuer
}

// Creates synchronous event emitter
func NewEvent[T any, E swarm.EventKind[T]](q *kernel.Enqueuer, category ...string) *EmitterEvent[T, E] {
	cat := swarm.TypeOf[E](category...)

	return &EmitterEvent[T, E]{
		cat:    cat,
		codec:  encoding.NewCodecEvent[T, E](q.Config.Source, cat),
		kernel: q,
	}
}

// Synchronously enqueue event to broker.
// It guarantees event to be send after return.
func (q *EmitterEvent[T, E]) Enq(ctx context.Context, object *E, cat ...string) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	category := q.cat
	if len(cat) > 0 {
		category = cat[0]
	}

	bag := swarm.Bag{
		Ctx:    swarm.NewContext(context.Background(), category, ""),
		Object: msg,
	}

	err = q.kernel.Emitter.Enq(ctx, bag)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

// Synchronous emitter of raw packets (bytes) to the broker.
// It blocks the routine until the message is accepted by the broker.
type EmitterBytes struct {
	cat    string
	codec  kernel.Encoder[[]byte]
	kernel *kernel.Enqueuer
}

// Creates synchronous emitter
func NewBytes(q *kernel.Enqueuer, category string) *EmitterBytes {
	var codec swarm.Codec
	if q.Config.Codec != nil {
		codec = q.Config.Codec
	} else {
		codec = encoding.NewCodecByte()
	}

	return &EmitterBytes{
		cat:    category,
		codec:  codec,
		kernel: q,
	}
}

// Synchronously enqueue bytes to broker.
// It guarantees message to be send after return
func (q *EmitterBytes) Enq(ctx context.Context, object []byte, cat ...string) error {
	msg, err := q.codec.Encode(object)
	if err != nil {
		return err
	}

	category := q.cat
	if len(cat) > 0 {
		category = cat[0]
	}

	bag := swarm.Bag{
		Ctx:    swarm.NewContext(context.Background(), category, ""),
		Object: msg,
	}

	err = q.kernel.Emitter.Enq(ctx, bag)
	if err != nil {
		return err
	}

	return nil
}
