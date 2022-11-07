
QoS: Message delivery guarantees

At most once:
Often called “best-effort”, where a message is published without any formal acknowledgement of receipt, and isn’t replayed.

Depending on the reliability of the network or system (as a whole), some messages can be lost as subscribers are not required to acknowledge receipt.

sending:
channel + dlq channel => buffered channels

receive:
receive + ack => buffered channels


At least once:
Typically implemented through a handshake or “acknowledgement” protocol, a message will be re-sent until it is formally acknowledged by a recipient.

A message can, depending on the behavior and configuration of the system, be re-sent and thus delivered more than once. Incurs a small overhead due to additional acknowledgement packets.

Subscribers can often handle duplicates at the persistency layer by ensuring each message carries a unique identifier or “idempotency key.” Even if the message is received more than once, the database layer will reject the duplicate key.

sending:
Enq() error call
unbuffered channel + dlq

receive:
receive + ack => unbuffered channel

Exactly once: 
It requires not only a way to acknowledge delivery, but additional state on the sender and receiver to ensure that a message is only accepted once, and that duplicates are discarded.


q := sqs.New(/* policy */)

snd, dlq := swarm.Enqueue[*User](q)
emitter  := swarm.Emitter[*User](q) 

rcv, ack := queue.Dequeue[*User](q)

selective receive:
eventbridge -> ok


//
//

sqs.New() -> Socket

type Socket interface {
  End(swarm.Bag) error
  Deq() (swarm.Bag, error)
  Ack(swarm.Bag) error
}

type Queue struct {
  socket Socket
  chans  map[]Closer  
}

func Close()

//
//
swarm.Enqueue[T any](Socket) (chan, chan)

//
queue.Enq
queue/event.Enq
queue/bytes.Enq

// terminology

// Canonical IRI that defines a type of action.
Type curie.IRI `json:"@type,omitempty"`
Category / Type of object

//
// Direct performer of the event, a software service that emits action to the stream.
Agent curie.IRI `json:"agent,omitempty"`

Event Bridge | Source = Agent = SQS | Agent 

//
// Indirect participants, a user who initiated an event.
Participant curie.IRI `json:"participant,omitempty"`

