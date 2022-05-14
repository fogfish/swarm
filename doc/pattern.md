# Design Pattern: Distributed event-driven Golang channels

Asynchronous communication is hard to implement but it is best reflection about properties of real world. Pure Go channels is a right abstraction for developers because it distills asynchronous semantic of communication into the native Golang code. The patter defines a fundamental idea how to use Golang channels to represents distributed environment as system of queues. 


## Context

Queuing systems or event brokers assist distributed solution with an asynchronous communication pattern between multiple processes/threads and other actors. These async patterns imply that execution of sender actor is not blocked while waiting response from recipient. Queuing systems "store-and-forward" enqueued messages until recipient retrieves them. It is important to emphasis, asynchronous actors delegate the responsibility of the message delivery to a queuing system, actors do not trace the status of the message delivery or expect immediate response from recipient. Instead, they expect particular grade of service (service level objectives) about messages delivery promise, asynchronous actors are always prepared for message loss and aspects of error handling.


## Problem 

Wide adoption of HTTP, which offers synchronous client-server communication patterns, has made a "negative" impact on message queues. Existing solution provides "wrong" abstraction for developers. It uses synchronous primitives to reflect pure asynchronous nature of queueing systems. As a result, engineers are forced to implement vendor dependent integrations or onboard layers of abstractions, which impacts on the maintainability and evolution of applications.  

Asynchronous communication is hard to implement but it is best reflection about properties of real world. Anything else is unnecessary complication that misleads developers by forcing usage of "wrong" interfaces that hides asynchronous nature behind vendor specific plumbing. The real solution is to write pure Go code instead of vendor specific APIs


## Solution

Pure Go channels is a right abstraction of queuing systems for developers because it distills asynchronous semantic of enqueueing/dequeueing into the native Golang code:

Sending messages are no more difficult than sending application specific instance of data type (Golang struct) to the channel
```go
queue <- Note{ID: "", Text: "some message"}
```

Receiving messages becomes a pure Golang idiomatic computation
```go
for msg := range queue {
  // ...
}
```

Usage of this abstraction improves non-functional requirements of your application:

1. **readability**: application uses pure Go code instead of vendor specific interfaces (learning time) 
2. **portability**: application is portable between various queuing systems in same manner as sockets abstracts networking stacks (exchange queueing transport "on-the-fly" to resolve evolution of requirements)  
3. **testability**: unit testing focuses on pure biz logic, simplify dependency injections and mocking (pure unit tests).  
4. **distribution**: idiomatic architecture to build distributed topologies and scale-out Golang applications (clustering).

## How it works

### Async I/O 

The fundamental idea of the pattern is that the pair of channels represents an instance of queue:
1. `<-chan T` channel to receive messages of type T 
2. `chan<- T` channel to send messages of type T 

The queue consumer becomes a pure Golang function either using `range` or `select` statements to receive messages: 

```go
func consume[T any](q <-chan T) {
  for msg := range q {
    // ...
  }
}
```

Queue clients become universal factory functions that construct the channels. Behind scene I/O on these channels are reflected to queue specific interface. 

```go
// 1. Instantiate client to message queue (transport layer)
q := sqs.Must(sqs.New(/* ... */))

// 2. Spawn Golang channel to receive messages from the queue
ch, _ := queue.Dequeue[Note](q)

// 3. Consume messages from channel
go consume(ch)
```

### Type safe logical channels

Let's consider an asynchronous communication scenario where application publishes messages in the format of JSON object. Application domain types are serialized as JSON messages with following exchange using queueing system. It immediately begs the question: Should the application share a single message queue among various types of business transactions or spawn a dedicated queue per type? Unfortunately, there are not a straight forward answer on the question. Each application makes decision based on various requirements such as operation management, separation of concerns, service-level objectives, high-availability, release cycles, etc.

Furthermore, usage of type safe development technique demands a message routing based on type of business transactions. In this respect, a queue abstraction needs to distinguish a layer of physical communication built either over octet streams; and a layer of logical communications that multiplexes types of business transactions over the physical layer

Queueing system uses concept of message attributes (metadata) to fulfil application's multiplexing requirements. Usage of these attributes to annotate the message for routing purpose is always a recommended despite application uses shared or dedicated queues. The instances of Golang channels are always dedicated to a domain type, represents a logical channel and abstracts the multiplexing complexity from developers: 

```go
sys := sqs.NewSystem("system-of-queues")
q := sqs.Must(sqs.New(sys, "instance-of-queue"))

// 
in, _ := queue.Dequeue[User](q)
eg, _ := queue.Enqueue[User](q)
```

Any message passed through these channels is annotated with following attributes:
* **System** is the logical name of the queueing system used to originate the message.
* **Queue** is the logical name of the queue used to originate the message.   
* **Category** is the name of category (data type).  

Applications build channels either communicate a `struct` of well-known type or slices of bytes `[]byte`, which opens an opportunity for the many encoding options like JSON or Gob.


### Fault tolerance and reliable communications

Usage of asynchronous communication patterns decouples fault tolerance concerns of sender and receiver -- actors do not need to interact with the message queue at the same time. Queuing systems ensures message delivery in the event of an actors failure. Naturally, service level objectives about message delivery varies between technologies and its operational domains but it remains to be opaque, actors cannot influence on it. However, the fault tolerant design demands both consumer and producer to implement protocols that
* acknowledges successful message processing - there is no guarantee in distributed system that the consumer receive and process message. Thus, the consumer must explicitly acknowledge message when it finishes processing.
* indicates failure of message transmission to messaging queue - following the statement above, messaging queue must explicitly acknowledge when it successfully scheduled the message for delivery. In case of unavailability, the message re-transmission is required. 

This protocol ensures that messages do not get "lost" in the event of a failure. 

Golang channel also facilitates implementation of this protocol. 

The event ingress part instantiates pair of Golang channels. The first readonly channel is used to process incoming messages, the second write-only channel is used to acknowledge successful processing by relaying incoming message back to queue.

```go
go consume(queue.Dequeue[User](q))

func consume(rcv <-chan *swarm.Msg[User], ack chan<- *swarm.Msg[User]) {
  for msg := range q {
    // ...
    ack <- msg
  }
}
```

The event egress part instantiates pair of Golang channels. The first write-only channel is used to enqueue message into the queue. The second readonly channel indicates failed transmissions, acting as dead-letter queue. Consumption of dead-letter queue helps application to recover from failure by retransmitting messages.

```go
go resend(queue.Enqueue[User](q))

func resend(out chan<- User, dlq <-chan User) {
  for msg := range dlq {
    // ...
    out <- msg
  }
}
```

## Consequences

### How to ensure order of events?

Golang channels guarantees the order of event only for messages of same kind (labelled with same System, Queue and Category). It preserve orders in the distributed scale only if underlying queueing system or event broker preserve ordering.

### Are there guarantees of the message delivery?

The message delivery is guaranteed by the underlying queueing system or event broker. Once the messages is accepted by the "transport", it is delivered according to promised grade of service. However, Golang channels are sophisticated "in-memory buffers". The channels introduces a lag of few milliseconds between scheduling a message to the channel and dispatching message for the processing by the consumer. There are two asymmetric channels:
* The dequeue channels guarantee reliable delivery of the message through the acknowledgement.  
* The enqueue channels are reliable in the absence of catastrophic failures. It's built-in retry logic protects losses from temporary unavailability of queueing transport. However, system might loss all the in-the-flight messages, thus the best effort delivery is guaranteed. The application shall use a synchronous primitives to enqueue messages if guaranteed delivery is required. 

### Can messages be delivered more than once?

Golang channels delivers message only once but overall property depends on the underlying queueing system or event broker.


## When to Use It

The pattern helps to address problems of application portability between different queueing systems, event brokers, vendors, etc especially in serverless environment. 

Also, use the pattern to build distributed systems with pure Go code instead of vendor specific interfaces.

## Examples

See [examples](../examples/) of usage Golang channels with various queueing systems.
