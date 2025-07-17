package websocket

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/curie/v2"
	"github.com/fogfish/golem/optics"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/kernel/encoding"
)

type EventCodec[E swarm.Event[M, T], M, T any] struct {
	encoding.Event[M, T]
	sink optics.Lens[M, curie.IRI]
}

func ForEvent[E swarm.Event[M, T], M, T any](realm, agent string, category ...string) EventCodec[E, M, T] {
	return EventCodec[E, M, T]{
		Event: encoding.ForEvent[E](realm, agent, category...),
		sink:  optics.ForProduct1[M, curie.IRI]("Sink"),
	}
}

func (c EventCodec[E, M, T]) Encode(evt swarm.Event[M, T]) (swarm.Bag, error) {
	bag, err := c.Event.Encode(evt)
	if err != nil {
		return swarm.Bag{}, err
	}

	if evt.IOContext != nil {
		switch ctx := evt.IOContext.(type) {
		case *events.APIGatewayWebsocketProxyRequestContext:
			bag.Category = ctx.ConnectionID
			return bag, nil
		}
	}

	if evt.Meta != nil {
		bag.Category = string(c.sink.Get(evt.Meta))
		return bag, nil
	}

	return bag, nil
}

func EmitEvent[E swarm.Event[M, T], M, T any](q *kernel.EmitterIO, codec ...kernel.Encoder[swarm.Event[M, T]]) (snd chan<- swarm.Event[M, T], dlq <-chan swarm.Event[M, T]) {
	if len(codec) > 0 {
		return kernel.EmitEvent(q, codec[0])
	}

	c := ForEvent[E](q.Config.Realm, q.Config.Agent)
	return kernel.EmitEvent(q, c)
}
