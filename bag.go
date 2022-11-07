package swarm

/*

Msg is a generic envelop type for incoming messages.
It contains both decoded object and its digest used to acknowledge message.
*/
type Msg[T any] struct {
	Object T
	Digest string
}

/*

Bag is an abstract container for octet stream.
Bag is used by the transport to abstract message on the wire.
*/
type Bag struct {
	Category string
	Object   []byte
	Digest   string
}
