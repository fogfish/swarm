<p align="center">
  <img src="./doc/swarm-v2.png" height="256" />
  <h3 align="center">swarm</h3>
  <p align="center"><strong>Go channels for queueing and event-driven systems</strong></p>

  <p align="center">
    <!-- Documentation -->
    <a href="http://godoc.org/github.com/fogfish/swarm">
      <img src="https://godoc.org/github.com/fogfish/swarm?status.svg" />
    </a>
    <!-- Build Status  -->
    <a href="https://github.com/fogfish/swarm/actions/">
      <img src="https://github.com/fogfish/swarm/workflows/build/badge.svg" />
    </a>
    <!-- GitHub -->
    <a href="http://github.com/fogfish/swarm">
      <img src="https://img.shields.io/github/last-commit/fogfish/swarm.svg" />
    </a>
    <!-- Coverage -->
    <a href="https://coveralls.io/github/fogfish/swarm?branch=main">
      <img src="https://coveralls.io/repos/github/fogfish/swarm/badge.svg?branch=main" />
    </a>
    <!-- Go Card -->
    <a href="https://goreportcard.com/report/github.com/fogfish/swarm">
      <img src="https://goreportcard.com/badge/github.com/fogfish/swarm" />
    </a>
    <!-- Maintainability -->
    <a href="https://codeclimate.com/github/fogfish/swarm/maintainability">
      <img src="https://api.codeclimate.com/v1/badges/6d525662ecccc2b9ff04/maintainability" />
    </a>
  </p>
</p>

---

Today's wrong abstractions lead to complexity on maintainability in the future. Usage of synchronous interfaces to reflect asynchronous nature of messaging queues is a good example of inaccurate abstraction. Usage of pure Go channels is a proper solution to distills asynchronous semantic of queueing systems into the idiomatic native Golang code.

## Inspiration

The library encourages developers to use Golang struct for asynchronous communication, which helps engineers to define domain models, write correct, maintainable code. The library uses generic programming style to abstract queueing systems into the idiomatic Golang channels `chan<- T` and `<-chan T`. See the design pattern [Distributed event-driven Golang channels](./doc/pattern.md) to learn philosophy and use-cases:
1. **readability**: application uses pure Go code instead of vendor specific interfaces (learning time) 
2. **portability**: application is portable between various queuing systems or event brokers in same manner as sockets abstracts networking stacks (exchange queueing transport "on-the-fly" to resolve evolution of requirements)  
3. **testability**: unit testing focuses on pure biz logic, simplify dependency injections and mocking (pure unit tests).  
4. **distribution**: idiomatic architecture to build distributed topologies and scale-out Golang applications (clustering).
5. **serverless**: of-the-shelf portable patterns for serverless applications (infrastructure as a code, aws cdk).

## Getting started

The library requires **Go 1.18** or later due to usage of [generics](https://go.dev/blog/intro-generics).

The latest version of the library is available at `main` branch of this repository. All development, including new features and bug fixes, take place on the `main` branch using forking and pull requests as described in contribution guidelines. The stable version is available via Golang modules.

Use `go get` to retrieve the library and add it as dependency to your application.

```bash
go get -u github.com/fogfish/swarm
```

- [Getting Started](#getting-started)
  - [Produce (enqueue) messages](#produce-enqueue-messages)
  - [Consume (dequeue) messages](#consume-dequeue-messages)
  - [Message Delivery Guarantees](#message-delivery-guarantees)
  - [Order of Messages](#order-of-messages)
  - [Octet streams](#octet-streams)
  - [Generic events](#generic-events)
  - [Serverless](#serverless)
  - [Supported queuing system and event brokers](#supported-queuing-system-and-event-brokers)
  <!-- - [Error handling](#error-handling) -->

### Produce (enqueue) messages

Please [see and try examples](examples). Its cover all basic use-cases with runnable code snippets, check the design pattern [Distributed event-driven Golang channels](./doc/pattern.md) for deep-dive into library philosophy.

The following code snippet shows a typical flow of sending the messages using the library.

```go
import (
  "github.com/fogfish/swarm/broker/sqs"
  "github.com/fogfish/swarm/queue"
)

// Use pure Golang struct to define semantic of messages and events
type Note struct {
  ID   string `json:"id"`
  Text string `json:"text"`
}

// Spawn a new instance of the messaging broker
q := swarm.Must(sqs.New("name-of-the-queue"))

// creates pair Golang channels dedicated for publishing
// messages of type Note through the messaging broker. The first channel
// is dedicated to emit messages. The second one is the dead letter queue that
// contains failed transmissions. 
enq, dlq := queue.Enqueue[*Note](q)

// Enqueue message of type Note
enq <- &Note{ID: "note", Text: "some text"}

// Close the broker and release all resources
q.Close()
```

### Consume (dequeue) messages

Please [see and try examples](examples). Its cover all basic use-cases with runnable code snippets, check the design pattern [Distributed event-driven Golang channels](./doc/pattern.md) for deep-dive into library philosophy.

The following code snippet shows a typical flow of sending the messages using the library.

```go
import (
  "github.com/fogfish/swarm/broker/sqs"
  "github.com/fogfish/swarm/queue"
)

// Use pure Golang struct to define semantic of messages and events
type Note struct {
  ID   string `json:"id"`
  Text string `json:"text"`
}

// Spawn a new instance of the messaging broker
q := swarm.Must(sqs.New("name-of-the-queue"))

// Create pair Golang channels dedicated for consuming
// messages of type Note from the messaging broker. The first channel
// is dedicated to receive messages. The second one is the channel to
// acknowledge consumption  
deq, ack := queue.Dequeue[Note](q)

// consume messages and then acknowledge it
for msg := range deq {
  ack <- msg
}

// Await messages from the broker
q.Await()
```

### Message Delivery Guarantees

Usage of Golang channels as an abstraction raises a concern about grade of service on the message delivery guarantees provided by the library. The library ensures exactly same grade of service as the underlying queueing system or event broker. It is delivered according to its promise once the messages is accepted by the remote peer of queuing system. The library's built-in retry logic protects losses from temporary unavailability of the remote peer. However, Golang channels are sophisticated "in-memory buffers", which introduce a lag of few milliseconds between scheduling a message to the channel and dispatching message to the remote peer. Use one of the following policy to either accept or protect from the loss all the in-the-flight messages in case of catastrophic failures.

#### At Most Once

**At Most Once** is best effort policy, where a message is published without any formal acknowledgement of receipt, and it isn't replayed. Some messages can be lost as subscribers are not required to acknowledge receipt.  

The library implements asymmetric approaches. The **enqueue** path uses buffered Golang channels for emitter and dead-letter queues. The **dequeue** path also uses buffered Golang channels for delivery message to consumer. The messages are automatically acknowledged to the broker upon successful scheduling. This means that information will be lost if the consumer crashes before it has finished processing the message. 

#### At Least Once

**At Least Once** is the default policy used by the library. The policy assume usage of "acknowledgement" protocol, which guarantees a message will be re-sent until it is formally acknowledged by a recipient. Messages should never be lost but it might be delivered more than once causing duplicate work to consumer.  

The library implements also asymmetric approaches. The **enqueue** path uses unbuffered Golang channels to emit messages and handle dead-letter queue, which leads to a delayed guarantee. The delayed guarantee in this context implies that enqueueing of other messages is blocked until dead-letter queue is resolved. Alternatively, the application can use synchronous protocol to enqueue message. The **dequeue** path also uses unbuffered Golang channels for delivery message to consumer and acknowledge its processing. The acknowledgement of message by consumer guarantee reliable delivery of the message but might cause duplicates. 

#### Exactly Once

Not supported by the library

### Order of Messages

TBD

### Octet Streams

The library support slices of bytes `[]byte` as message type. It opens an opportunity for the many encoding options like JSON, Gob, etc.  

```go
import (
  queue "github.com/fogfish/swarm/queue/bytes"
)

enq, dlq := queue.Enqueue(q, "Note")
deq, ack := queue.Dequeue(q, "Note")
```

Please see example about [binary](./examples/bytes) consumer/producer.

### Generic events

Event defines immutable fact(s) placed into the queueing system.
Event resembles the concept of [Action](https://schema.org/Action) as it is defined by schema.org.

> An action performed by a direct agent and indirect participants upon a direct object.

This type supports development of event-driven solutions that treat data as
a collection of immutable facts, which are queried and processed in real-time.
These applications processes logical log of events, each event defines a change
to current state of the object, i.e. which attributes were inserted,
updated or deleted (a kind of diff). The event identifies the object that was
changed together with using unique identifier.

The library support this concept through generic type `swarm.Event[T]`.  

```go
import (
  "github.com/fogfish/swarm"
  queue "github.com/fogfish/swarm/queue/events"
)

type EventCreateNote swarm.Event[*Note]

func (EventCreateNote) HKT1(swarm.EventType) {}
func (EventCreateNote) HKT2(*Note)           {}

enq, dlq := queue.Enqueue[*Note, EventCreateNote](q)
deq, ack := queue.Dequeue[*Note, EventCreateNote](q)
```

Please see example about [event-driven](./examples/events/) consumer/producer.

### Serverless 

The library support development of serverless event-driven application using AWS service. The library provides AWS CDK Golang constructs to spawn consumers. See example of [serverless consumer](./examples/eventbridge/recv/eventbridge.go) and corresponding AWS CDK [application](./examples/eventbridge/serverless/main.go).

```go
package main

import (
  "github.com/fogfish/scud"
  "github.com/fogfish/swarm/queue/eventbridge"
)

func main() {
  app := eventbridge.NewServerlessApp()

  stack := app.NewStack("swarm-example-eventbridge")
  stack.NewEventBus()

  stack.NewSink(
    &eventbridge.SinkProps{
      Agents: []string{"swarm-example-eventbridge"},
      Lambda: &scud.FunctionGoProps{
        SourceCodePackage: "github.com/fogfish/swarm",
        SourceCodeLambda:  "examples/eventbridge/recv",
      },
    },
  )

  app.Synth(nil)
}
```

### Supported queuing system and event brokers 

- [x] AWS EventBridge (serverless only)
  - [x] [sending message](examples/eventbridge/send/eventbridge.go)
  - [x] [receiving message](examples/eventbridge/recv/eventbridge.go) using aws lambda
  - [x] [aws cdk construct](examples/eventbridge/serverless/main.go)
- [x] AWS SQS (serverless)
  - [x] [sending message](examples/eventsqs/send/eventsqs.go)
  - [x] [receiving message](examples/eventsqs/recv/eventsqs.go) using aws lambda
  - [x] [aws cdk construct](examples/eventsqs/serverless/main.go)
- [x] AWS SQS
  - [x] [sending message](examples/sqs/send/sqs.go)
  - [x] [receiving message](examples/sqs/recv/sqs.go)
- [ ] AWS SNS
  - [ ] sending message
- [ ] AWS Kinesis (serverless)
  - [ ] sending message
  - [ ] receiving message using aws lambda
- [ ] AWS Kinesis
  - [ ] sending message
  - [ ] receiving message
- [ ] Redis
  - [ ] sending message
  - [ ] receiving message
- [ ] MQTT API
  - [ ] sending message
  - [ ] receiving message

Please let us know via [GitHub issues](https://github.com/fogfish/swarm/issue) your needs about queuing technologies.


## How To Contribute

The library is [Apache Version 2.0](LICENSE) licensed and accepts contributions via GitHub pull requests:

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

The build and testing process requires [Go](https://golang.org) version 1.16 or later.

**build** and **test** library.

```bash
git clone https://github.com/fogfish/swarm
cd swarm
go test ./...
```

### commit message

The commit message helps us to write a good release note, speed-up review process. The message should address two question what changed and why. The project follows the template defined by chapter [Contributing to a Project](http://git-scm.com/book/ch5-2.html) of Git book.

### bugs

If you experience any issues with the library, please let us know via [GitHub issues](https://github.com/fogfish/swarm/issue). We appreciate detailed and accurate reports that help us to identity and replicate the issue. 

### benchmarking

```bash
cd queue/sqs
go test -run=^$ -bench=. -benchtime 100x
```


## Bring Your Own Queue

TBD

## License

[![See LICENSE](https://img.shields.io/github/license/fogfish/swarm.svg?style=for-the-badge)](LICENSE)

