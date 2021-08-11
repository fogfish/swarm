# Design Pattern: Message Queues in Golang

TODO: Intent 


## Context

Message Queues assists distributed systems with an asynchronous communication pattern between multiple processes/threads aka actors. These async patterns imply that execution of sender actor is not blocked while waiting response from recipient. Instead queuing systems "store-and-forward" enqueued messages until recipient retrieves them. It is important to emphasis, asynchronous actors delegate the responsibility of the message delivery to a queuing system, actors do not trace the status of the message delivery or expect immediate response from recipient. Instead, they expect particular grade of service (service level objectives) about messages delivery promise, asynchronous actors are always prepared for message loss and aspects of error handling.


## Problem 

Wide adoption of HTTP, which offers synchronous client-server communication patterns, has made a "negative" impact on message queues. Existing solution provides "wrong" abstraction for developers. It uses synchronous primitives to reflect asynchronous nature of queueing systems. As the result, ... TODO ...

... (layers of abstractions, plumbing, vendor dependent integrations, maintainability and evolution) ...

Asynchronous communication is hard to implement but it is best reflection about properties of real world.

anything else is unnecessary complication that misleads developers by forcing usage of sync API primitives. (Hiding async aspects behind vendor api plumbing). Write idiomatic Go code instead of plumbing queue specific APIs


## Solution

Pure Go channels is a right abstraction of queuing systems for developers because it distills asynchronous semantic of enqueueing/dequeueing into the idiomatic native Golang code:

Sending messages are no more difficult than
```go
queue <- []byte("some message")
```

Receiving messages becomes a pure Golang idiomatic computation
```go
for msg := range queue {
  // ...
}
```

Usage of this abstraction improves non-functional requirements of your application:

1. readability: application uses idiomatic Go code instead of vendor specific interfaces (learning time) 
2. portability: application is portable between various queuing systems in same manner as sockets abstracts networking stacks. (exchange tech stack, Develop against an in-memory)
3. testability: (unit tests with no dependencies)


## How it works

### Async I/O 

The fundamental idea of the pattern is that the pair of channels represents an instance of queue:
1. `<-chan []byte` channel to receive messages
2. `chan<- []byte` channel to send messages 

The queue consumer becomes a pure Golang function either using `range` or `select` statements to receive messages: 

```go
func consume(q <-chan []byte) {
  for msg := range q {
    // ...
  }
}
```

Queue clients become universal factory functions that construct the channels. Behind scene I/O on these channels are reflected to queue specific interface. 

```go
in, out := queue.New(/* ... */)

go consume(in)
```

### Type safe logical channels

Let's consider an asynchronous communication scenario where application publishes messages in the format of JSON object. Application domain types are serialized as JSON messages with following exchange using queueing system. It immediately begs the question: Should the application share a single message queue among various types of business transactions or spawn a dedicated queue per type? Unfortunately, there are not a straight forward answer on the question. Each application makes decision based on various requirements such as operation management, separation of concerns, service-level objectives, high-availability, release cycles, etc.

Furthermore, usage of type safe development technique demands a message routing based on type of business transactions. In this respect, a queue abstraction needs to distinguish a layer of physical communication built either over octet streams; and a layer of logical communications that multiplexes types of business transactions over the physical layer

Queueing system uses concept of message attributes (metadata) to fulfil application's multiplexing requirements. Usage of these attributes to annotate the message for routing purpose is always a recommended despite application uses shared or dedicated queues.

The pair of channels is dedicated to a domain type, represents a logical channel and abstracts the multiplexing complexity from developers: 

```go
q := queue.New(/* ... */)

rcv := q.Recv(/* type identifier */)
snd := q.Send(/* type identifier */)
```

These channels either communicate a `struct` of well-known type or slices of bytes `[]byte`, which opens an opportunity for the many encoding options like JSON or Gob.


### Fault tolerance

Usage of asynchronous communication patterns decouples fault tolerance concerns of sender and receiver -- actors do not need to interact with the message queue at the same time. Queuing systems ensures message delivery in the event of an actors failure. Naturally, service level objectives about message delivery varies between technologies and its operational domains but it remains to be opaque, actors cannot influence on it. However, the fault tolerant design demands both sender and receiver to implement a protocol that
* acknowledges successful message processing - there is no guarantee in distributed system that the consumer receive and process message. Thus, the consumer must explicitly acknowledge message when it finishes processing.
* indicates failure of message transmission to messaging queue - following the statement above, messaging queue must explicitly acknowledge when it successfully scheduled the message for delivery. In case of unavailability, the message re-transmission is required. 

This protocol ensures that messages do not get "lost" in the event of a failure. 

Golang channel also facilitates implementation of this protocol. 

...TODO...
```go
go consume(q.Recv(/* type identifier */))

func consume(rcv <-chan []byte, ack chan<- []byte) {
  for msg := range q {
    // ...
    ack <- msg
  }
}
```

...TODO...
```go
go resend(q.Send(/* type identifier */))

func resend(snd chan<- []byte, err <-chan []byte) {
  for msg := range err {
    // ...
    snd <- msg
  }
}
```

## When to Use It



## Examples
