package swarm

/*

Msg type defines external ingress message.
It containers both payload and receipt to acknowledge
*/
type Msg struct {
	Payload []byte
	Receipt string
}

var (
	_ MsgV0 = (*Msg)(nil)
)

/*

Bytes returns message payload (octet stream)
*/
func (msg *Msg) Bytes() []byte {
	return msg.Payload
}

/*

Bag is an internal message envelop containing message and routing attributes

TODO: Identity id
 - System
 - Actor
 - Category (type)
 - ==> Or category:system/actor

*/
type Bag struct {
	// routing attributes
	Target   string
	Source   string
	Category Category

	// message payload
	Object MsgV0

	//
	StdErr chan<- MsgV0
}
