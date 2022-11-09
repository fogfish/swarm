package swarm

type Broker interface {
	Config() *Config
	Close()
	Await()
	Enqueue(string, Channel) (Enqueue, error)
	Dequeue(string, Channel) (Dequeue, error)
}

type Enqueue interface {
	Enq(Bag) error
}

type Dequeue interface {
	Deq(string) (Bag, error)
	Ack(Bag) error
}
