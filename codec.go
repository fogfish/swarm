package swarm

import (
	"encoding/json"
	"time"

	"github.com/fogfish/curie"
	"github.com/fogfish/golem/optics"
	"github.com/fogfish/guid/v2"
)

//------------------------------------------------------------------------------

// Json codec for I/O kernel
type CodecJson[T any] struct{}

func (CodecJson[T]) Encode(x T) ([]byte, error) {
	return json.Marshal(x)
}

func (CodecJson[T]) Decode(b []byte) (x T, err error) {
	err = json.Unmarshal(b, &x)
	return
}

func NewCodecJson[T any]() CodecJson[T] { return CodecJson[T]{} }

//------------------------------------------------------------------------------

// Byte identity codec for I/O kernet
type CodecByte struct{}

func (CodecByte) Encode(x []byte) ([]byte, error) { return x, nil }
func (CodecByte) Decode(x []byte) ([]byte, error) { return x, nil }

func NewCodecByte() CodecByte { return CodecByte{} }

//------------------------------------------------------------------------------

// Event codec for I/O kernel
type CodecEvent[T any, E EventKind[T]] struct {
	source string
	cat    string
	shape  optics.Lens4[E, string, curie.IRI, curie.IRI, time.Time]
}

func (c CodecEvent[T, E]) Encode(obj *E) ([]byte, error) {
	_, knd, src, _ := c.shape.Get(obj)
	if knd == "" {
		knd = curie.IRI(c.cat)
	}

	if src == "" {
		src = curie.IRI(c.source)
	}

	c.shape.Put(obj, guid.G(guid.Clock).String(), knd, src, time.Now())

	return json.Marshal(obj)
}

func (c CodecEvent[T, E]) Decode(b []byte) (*E, error) {
	x := new(E)
	err := json.Unmarshal(b, x)

	return x, err
}

func NewCodecEvent[T any, E EventKind[T]](source, cat string) CodecEvent[T, E] {
	return CodecEvent[T, E]{
		source: source,
		cat:    cat,
		shape:  optics.ForShape4[E, string, curie.IRI, curie.IRI, time.Time]("ID", "Type", "Agent", "Created"),
	}
}
