# Design Pattern: Distributed event-driven Golang channels

Asynchronous communication is hard to implement. However, it is best reflection about properties of real world. Pure Go channels is a right abstraction for developers because it distills asynchronous semantic of communication into the native Golang code. This design patter defines a fundamental idea how to use Golang channels to represents distributed environment as system of queues. 


## Context

Queuing systems and Event brokers assist distributed solutions with an asynchronous communication pattern between multiple actors - processes/threads, producers/consumers, etc. These async patterns imply that execution of producer is not blocked while waiting response from recipient. Queuing systems "store-and-forward" enqueued messages until recipient retrieves them. It is important to emphasis that asynchronous actors delegate the responsibility of the message delivery to a queuing system. Actors do not trace the status of the message delivery or expect immediate response from recipient. Instead, they expect particular grade of service (service level objectives) about messages delivery promise.


## Problem 

Wide adoption of HTTP(S), which offers synchronous client-server communication patterns, has made a "negative" impact on message queues. Existing solution provides "wrong" abstraction for developers. It uses synchronous primitives to reflect pure asynchronous nature of queueing systems. As a result, engineers are forced to implement vendor dependent integrations or onboard layers of abstractions, which impacts on the maintainability and evolution of applications.  

Asynchronous communication is hard to implement but it is best reflection about properties of real world. Anything else is unnecessary complication that misleads developers by forcing usage of "wrong" interfaces that hides asynchronous nature behind vendor specific plumbing. The real solution is to write pure Go code instead of vendor specific APIs


## Solution

Pure Go channels is a right abstraction of queuing systems for developers because it distills asynchronous semantic of enqueueing/dequeueing into the native Golang code:

Sending messages are no more difficult than sending application specific instance of data type (Golang struct) to the channel
```go
queue <- Note{ID: "note", Text: "some message"}
```

Receiving messages becomes a pure Golang idiomatic computation
```go
for msg := range queue {
  // ...
}
```

Usage of this abstraction improves non-functional requirements of your application:

1. **readability**: application uses pure Go code instead of vendor specific interfaces and rest api (learning time) 
2. **portability**: application is portable between various queuing systems in same manner as sockets abstracts networking stacks (exchange queueing transport "on-the-fly" to resolve evolution of requirements)  
3. **testability**: unit testing focuses on pure biz logic, simplify dependency injections and mocking (pure unit tests).  
4. **distribution**: idiomatic architecture to build distributed topologies and scale-out Golang applications (clustering).

## How it works

### Basic Channels 

The fundamental idea of the pattern is that the pair of channels represents an instance of queue:
1. `<-chan T` channel to consume (dequeue) messages of type T 
2. `chan<- T` channel to produce (enqueue) messages of type T 

The consumer becomes a pure Golang function either using `range` or `select` statements to receive messages: 

```go
func consume[T any](q <-chan T) {
  for msg := range q {
    // ...
  }
}
```

We are only missing the method to create these channels. The pattern defined two abstractions:
* `Broker` is the adapter of queuing systems. It defines synchronous primitives to spawn the new instance of client, enqueue, dequeue and acknowledge the message.   
* `Channel factory` is a collection of Golang routines that projects synchronous primitives into the channels behind scene.


```go
// Spawn a new instance of the messaging broker
q := sqs.New(/* ... */)

// Create Golang channel for consuming messages of type Note
// from the messaging broker 
deq := queue.Dequeue[Note](q)

// Consume messages from the channel
go consume(deq)
```

### Type Safe Channels

Let's consider an asynchronous communication scenario where application publishes messages in the format of JSON object. Application domain types are serialized as JSON messages with consequent exchange using queueing system. It immediately begs the question: Should the application share a single message queue among various types of business transactions or spawn a dedicated queue per type? Unfortunately, there are not a straight forward answer on the question. Each application makes decision based on various requirements such as operation management, separation of concerns, service-level objectives, high-availability, release cycles, etc.

Furthermore, usage of type safe development technique demands a message routing based on type of business transactions. In this respect, a queue abstraction needs to distinguish a layer of physical communication built either over octet streams; and a layer of logical communications that multiplexes types of business transactions over the physical layer

Queueing system uses concept of message attributes (metadata) to fulfil application's multiplexing requirements. Usage of these attributes to annotate the message for routing purpose is always a recommended despite application uses shared or dedicated queues. The instances of Golang channels are always dedicated to a domain type, represents a logical channel and abstracts the multiplexing complexity from developers: 

```go
q := sqs.New(/* ... */)

deq := queue.Dequeue[User](q)
enq := queue.Enqueue[User](q)
```

Any message passed through these channels is annotated with following attributes:
* **Agent** is a direct performer of the message, it is a software service that emits action to the queueing system.
* **Category** is the identity of data type the message belong to. 

Using these attributes, messages are not only routed through queuing system but also routed to designated type safe channel.
```go
user := queue.Dequeue[User](q)
note := queue.Dequeue[Note](q)
```

### Delivery Guarantees with Golang channels

Usage of Golang channels as an abstraction raises a concern about grade of service on the message delivery guarantees. The pattern ensures exactly same grade of service as the underlying queueing system. It is delivered according to its promise once the messages is accepted by the remote peer of queueing system.The pattern assumes the retry logic to protect from losses due to temporary unavailability of the remote peer. However, Golang channels are sophisticated "in-memory buffers", which introduce a lag of few milliseconds between scheduling a message to the channel and dispatching message to the remote peer. The pattern assumes one of the following policy to either accept or protect from the loss all the in-the-flight messages in case of catastrophic failures.

#### At Most Once

**At Most Once** is best effort policy, where a message is published without any formal acknowledgement of receipt, and it isn't replayed. Some messages can be lost as subscribers are not required to acknowledge receipt. Usage of buffered Golang channels is right approach to implement this policy.   

#### At Least Once

**At Least Once** is policy assume usage of "acknowledgement" protocol, which guarantees a message will be re-sent until it is formally acknowledged by a recipient. Messages should never be lost but it might be delivered more than once causing duplicate work to consumer.  

The pattern proposes asymmetric approaches.
* acknowledges successful consumption of message - there is no guarantee in distributed system that the consumer receive and process message. Thus, the consumer must explicitly acknowledge message when it finishes processing.
* indicates failure of message transmission to messaging queue - following the statement above, messaging queue must explicitly acknowledge when it successfully scheduled the message for delivery. In case of unavailability, the message re-transmission is required. 

This protocol ensures that messages do not get "lost" in the event of a failure. 

Unbuffered Golang channel facilitates implementation of this protocol. 

The message consumer instantiates pair of Golang channels. The first readonly channel is used to accept incoming messages, the second write-only channel is used to acknowledge successful processing by relaying incoming message back to queue.

```go
go consume(queue.Dequeue[*Note](q))

func consume(deq <-chan *Note, ack chan<- *Note) {
  for msg := range deq {
    // ...
    ack <- msg
  }
}
```

The message producer instantiates pair of Golang channels. The first write-only channel is used to enqueue message into the queue. The second readonly channel indicates failed transmissions, acting as dead-letter queue. Consumption of dead-letter queue helps application to recover from failure by retransmitting messages.

```go
go resend(queue.Enqueue[User](q))

func resend(out chan<- User, dlq <-chan User) {
  for msg := range dlq {
    // ...
    out <- msg
  }
}
```

#### Exactly Once

**Exactly once** guarantees receive a given message once. Failures and retries may occur by they are invisible for consumer.

The pattern does not solve this policy using any of the sophisticated approaches. It uses a naive approach of deduplicating processed messages at the channel designated for consumer. 


## Consequences

### How to ensure order of events?

Golang channels guarantees the order of event only for messages of same kind (labelled with same Category). It preserve orders in the distributed scale only if underlying queueing system or event broker preserve ordering.

### Are there guarantees of the message delivery?

The message delivery is guaranteed by the underlying queueing system or event broker. Once the messages is accepted by the "transport", it is delivered according to promised grade of service. However, Golang channels are sophisticated "in-memory buffers". The channels introduces a lag of few milliseconds between scheduling a message to the channel and dispatching message for the processing by the consumer. There are two asymmetric channels:
* The dequeue channels guarantee reliable delivery of the message through the acknowledgement.  
* The enqueue channels are reliable in the absence of catastrophic failures. It's built-in retry logic protects losses from temporary unavailability of queueing transport. However, system might loss all the in-the-flight messages, thus the best effort delivery is guaranteed. The application shall use a synchronous primitives to enqueue messages if guaranteed delivery is required. 

### Can messages be delivered more than once?

Golang channels delivers message only once but overall property depends on the underlying queueing system or event broker.


## When to Use It

The pattern helps to address problems of application portability between different queueing systems, event brokers or vendors especially in serverless environment. 

Also, use the pattern to build distributed systems with pure Go code instead of vendor specific interfaces.

## Examples

See [examples](../examples/) of usage Golang channels with various queueing systems.
