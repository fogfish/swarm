package queue

// /*

// Msg type defines external ingress message.
// It containers both payload and receipt to acknowledge
// */
// type Msg struct {
// 	Payload []byte
// 	Receipt string
// }

// var (
// 	_ swarm.Msg = (*Msg)(nil)
// )

// /*

// Bytes returns message payload (octet stream)
// */
// func (msg *Msg) Bytes() []byte {
// 	return msg.Payload
// }

// /*

// Bag is an internal message envelop containing message and routing attributes

// TODO: Identity id
//  - System
//  - Actor
//  - Category (type)
//  - ==> Or category:system/actor

// */
// type Bag struct {
// 	// routing attributes
// 	Target   string
// 	Source   string
// 	Category swarm.Category

// 	// message payload
// 	Object swarm.Msg

// 	//
// 	StdErr chan<- swarm.Msg
// }

// /*

// bagRecv is the pair of channel, exposed by the transport queue to proxy
// */
// type bagRecv struct {
// 	msg <-chan *Bag
// 	ack chan<- *Bag
// }
