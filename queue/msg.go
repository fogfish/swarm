package queue

import "github.com/fogfish/swarm"

/*

Msg type defines external ingress message.
It containers both payload and receipt to acknowledge
*/
type Msg struct {
	Payload []byte
	Receipt string
}

var (
	_ swarm.Msg = (*Msg)(nil)
)

/*

Bytes returns message payload (octet stream)
*/
func (msg *Msg) Bytes() []byte {
	return msg.Payload
}

/*

Bag is a product type of message and its attributes
*/
type Bag struct {
	// message attributes
	Target   string
	Source   string
	Category swarm.Category

	// message payload
	Object swarm.Msg

	//
	StdErr chan<- swarm.Msg
}
