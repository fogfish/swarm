<p align="center">
  <img src="./swarm-v3.png" height="256" />
  <h3 align="center">swarm</h3>
  <p align="center"><strong>Go channels for distributed queueing and event-driven systems</strong></p>
  <h1 align="center">User Guide</h1>
</p>

`swarm` is a library that bridges the gap between Go's elegant concurrency model and the complex reality of distributed queues and event brokers.

It wraps popular messaging systems behind a **simple, type-safe interface** based on Go channels:

This resolves **the hidden complexity crisis**: developers build asynchronous architectures but are forced to work with synchronous interfaces, creating a mismatch that increases complexity and maintenance overhead.

Talk to channels - not brokers - to make your code clean, testable (with in-memory backends), and portable across messaging technologies. Transform your distributed systems into elegant Go code with `swarm`.

- [Why Go channels are perfect for distributed systems?](#why-go-channels-are-perfect-for-distributed-systems)
  - [The Mental Model Shift](#the-mental-model-shift)
- [Produce (emit) messages](#produce-emit-messages)
- [Consume (listen) messages](#consume-listen-messages)
- [Configure messaging broker](#configure-messaging-broker)
- [Message Delivery Guarantees](#message-delivery-guarantees)
- [Delayed Guarantee vs Guarantee](#delayed-guarantee-vs-guarantee)
- [Order of Messages](#order-of-messages)
- [Octet Streams](#octet-streams)
- [Generic events](#generic-events)
- [Error Handling](#error-handling)
- [Fail Fast](#fail-fast)
- [Serverless](#serverless)


## Why Go channels are perfect for distributed systems?

Go channels provide an ideal abstraction for distributed messaging because they already embody the correct mental model for asynchronous communication.

**Asynchronous Semantics**: Go channels are inherently asynchronous by design. Unlike traditional messaging APIs that often expose synchronous-looking interfaces (like send() and receive() methods), channels naturally represent the flow of messages between independent processes. This aligns perfectly with how distributed systems actually work.

**Familiar Concurrency Patterns**: Go developers already understand channel patterns like fan-in, fan-out, and pipeline processing. Swarm extends these familiar patterns to distributed systems, making it easier to reason about complex messaging topologies without learning new abstractions.

**Simplified Error Handling**: Channels provide a consistent error handling model through dead letter queues and acknowledgment channels. This eliminates the need to learn different error handling patterns for each messaging technology.

**Type Safety and Generic Programming**: Go's generic programming capabilities to provide type-safe messaging without reflection or runtime type checking. This means you can work with strongly-typed message structs.

With `swarm`, you don’t need to learn new abstractions - you just write idiomatic Go.

### The Mental Model Shift

`swarm` isn't just about replacing HTTP calls with message queues - it's about embracing the asynchronous nature of distributed systems from the ground up.

Traditional distributed systems often try to **make remote calls look like local function calls**. This creates false expectations about reliability, latency, and failure modes. `swarm` embraces the reality that **distributed communication is inherently asynchronous and unreliable**.

By using Go channels as the abstraction layer, `swarm` helps developers think in terms of **message flows rather than request-response cycles**. This leads to more resilient, scalable, and maintainable distributed systems.

The key insight is that **the right abstraction doesn't hide complexity - it makes complexity manageable**. `swarm` makes the asynchronous nature of distributed systems explicit through channels, while hiding the vendor-specific implementation details behind a familiar Go interface.


## Produce (emit) messages

The following code snippet shows a typical flow of producing the messages using the library.

```go
import (
  "github.com/fogfish/swarm/broker/sqs"
  "github.com/fogfish/swarm/emit"
)

// Use pure Golang struct to define semantic of messages and events
type Note struct {
  ID   string `json:"id"`
  Text string `json:"text"`
}

// Spawn a new instance of the messaging broker
q := sqs.Channels().MustClient("aws-sqs-queue-name")

// creates pair Golang channels dedicated for publishing
// messages of type [Note] through the messaging broker. The first channel
// is dedicated to emit messages. The second one is the dead letter queue that
// contains failed transmissions. 
enq, dlq := emit.Typed[Note](q)

// Enqueue message of type Note
enq <- Note{ID: "note", Text: "some text"}

// Close the broker and release all resources
q.Close()
```

## Consume (listen) messages

The following code snippet shows a typical flow of consuming the messages using the library.

```go
import (
  "github.com/fogfish/swarm/broker/sqs"
  "github.com/fogfish/swarm/listen"
)

// Use pure Golang struct to define semantic of messages and events
type Note struct {
  ID   string `json:"id"`
  Text string `json:"text"`
}

// Spawn a new instance of the messaging broker
q := sqs.Channels().MustClient("aws-sqs-queue-name")

// Create pair Golang channels dedicated for consuming
// messages of type Note from the messaging broker. The first channel
// is dedicated to receive messages. The second one is the channel to
// acknowledge consumption  
deq, ack := listen.Typed[Note](q)

// consume messages and then acknowledge it
go func() {
  for msg := range deq {
    /* ... do something with msg.Object and ack the message ...*/
    ack <- msg
  }
}()

// Await messages from the broker
q.Await()
```

## Configure messaging broker

The library uses the builder pattern to construct broker interfaces. Each broker exposes a `Channels()` method, which returns a broker-specific builder interface. This builder provides broker-specific options, including a `WithKernel(...)` method to configure a generic messaging kernel.

For details on kernel configuration, refer to [config.go](./config.go).

```go
sqs.Channels().
  WithBatchSize(5).
  WithKernel(
    swarm.WithSource("name-of-my-component"),
    swarm.WithRetryConstant(10 * time.Millisecond, 3),
    swarm.WithPollFrequency(10 * time.Second),
    /* ... */
  ).
  NewClient("name-of-the-queue")
```

## Message Delivery Guarantees

Usage of Golang channels as an abstraction raises a concern about grade of service on the message delivery guarantees. The library ensures exactly same grade of service as the underlying queueing system or event broker. Messages are delivered according to the promise once they are accepted by the remote side of queuing system. The library's built-in retry logic protects losses from temporary unavailability of the remote peer. However, Golang channels function as sophisticated "in-memory buffers," which can introduce a delay of a few microseconds between scheduling a message to the channel and dispatching it to the remote peer. To handle catastrophic failures, choose one of the following policies to either accept or safeguard in-flight messages from potential loss.

**At Most Once** is best effort policy, where a message is published without any formal acknowledgement of receipt, and it isn't replayed. Some messages can be lost as subscribers are not required to acknowledge receipt.  

The library implements asymmetric approaches for message handling. In the **emit** path, buffered Golang channels are used for both message emission and managing dead-letter queues. Similarly, the **listen** path uses buffered Golang channels to deliver messages to the consumer.   

```go
// Spawn a new instance of the messaging broker using At Most Once policy.
// The policy defines the capacity of Golang channel.
q := sqs.Channels().
  WithKernel(
    swarm.WithPolicyAtMostOnce(1000),
  ).
  MustClient("name-of-the-queue")

// emit channels has capacity 1000
enq, dlq := emit.Typed[Note](q)

// recv channels has capacity 1000
deq, ack := listen.Typed[Note](q)
```

**At Least Once** is the default policy used by the library. The policy assume usage of "acknowledgement" protocol, which guarantees a message will be re-sent until it is formally acknowledged by a recipient. Messages should never be lost but it might be delivered more than once causing duplicate work to consumer.  

The library also implements asymmetric approaches for message handling. In the **emit** path, unbuffered Golang channels are used to emit messages and manage the dead-letter queue, resulting in a delayed guarantee. This means that emitting additional messages is blocked until the dead-letter queue is fully resolved. Alternatively, the application can opt for a synchronous protocol to emit messages.

In the **listen** path, buffered Golang channels are used to deliver messages to the consumer and acknowledge their processing. While consumer acknowledgment ensures reliable message delivery, it may lead to message duplication.

```go
// Spawn a new instance of the messaging broker using At Least Once policy.
// At Least Once policy is the default one, no needs to explicitly declare it.
// Use it only if you need to define other capacity for listen channel than
// the default one, which creates unbuffered channel
q := sqs.Channels().
  WithKernel(
    swarm.WithPolicyAtLeastOnce(1000),
  ).
  MustClient("name-of-the-queue")

// both channels are unbuffered
enq, dlq := emit.Typed[Note](q)

// buffered channels of capacity n
deq, ack := listen.Typed[Note](q)
```

**Exactly Once** is not supported by the library yet.

// TODO: custom capacity.

## Delayed Guarantee vs Guarantee

Usage of **At Least Once** policy (unbuffered channels) provides the delayed guarantee for producers. Let's consider the following example. If queue broker fails to send message `A` then the channel `enq` is blocked at sending message `B` until the program consumes message `A` from the dead-letter queue channel.

```go
enq, dlq := emit.Typed[User](q)

enq <- User{ID: "A", Text: "some text by A"} // failed to send
enq <- User{ID: "B", Text: "some text by B"} // blocked until dlq is processed 
enq <- User{ID: "C", Text: "some text by C"}
```

The delayed guarantee is efficient on batch processing, pipelining but might cause complication at transactional processing. Therefore, the library also support a synchronous variant to producing a message:

```go
// Creates "synchronous" variant of the queue
user := emit.NewTyped[User](q)

// Synchronously emit the message. It ensure that message is scheduled for
// delivery to remote peer once function successfully returns.
if err := user.Enq(context.Background(), &User{ID: "A", Text: "some text by A"}); err != nil {
  // handle error
}
```

## Order of Messages

The library guarantee ordering of the messages when they are produced over same Golang channel. Let's consider a following example:

```go
user, _ := emit.Typed[User](q)
note, _ := emit.Typed[Note](q)

user <- &User{ID: "A", Text: "some text by A"}
note <- &Note{ID: "B", Text: "some note A"}
user <- &User{ID: "C", Text: "some text by A"}
```

The library guarantees following clauses `A before C` and `C after A` because both messages are produced to single channel `user`. It do not guarantee clauses `A before B`, `B before C` or `C after B` because multiple channels are used.

The library does not provide any higher guarantee than underlying message broker. For example, using SQS would not guarantee any ordering while SQS FIFO makes sure that messages of same type is ordered.


## Octet Streams

The library support slices of bytes `[]byte` as message type. It opens an opportunity for the many encoding options like JSON, Gob, etc.  

```go
import (
  "github.com/fogfish/swarm/emit"
  "github.com/fogfish/swarm/listen"
)

enq, dlq := emit.Bytes(q, codec)
deq, ack := listen.Bytes(q, codec)
```

Please see example about binary [producer](./broker/sqs/examples/emit/bytes/sqs.go) and [consumer](./broker/sqs/examples/listen/bytes/sqs.go).


## Generic events

An event represents an immutable fact placed into the queuing system. It is conceptually similar to the [Action](https://schema.org/Action) defined by schema.org.

> An action performed by a direct agent and indirect participants upon a direct object.

This type facilitates the development of event-driven solutions that treat data as a collection of immutable facts, which can be queried and processed in real-time. These applications process a logical event log, where each event represents a change to the current state of an object, such as which attributes were inserted, updated, or deleted (essentially a diff). Each event uniquely identifies the affected object using a unique identifier.

Unlike other solutions, this approach does not use an envelope for events. Instead, it pairs metadata and data side by side, making it more extendable.

```go

type Meta struct {
  swarm.Meta
  About string `json:"about"`
}

type User struct {
  ID   string `json:"id"`
  Text string `json:"text"`
}

type UserEvent = swarm.Event[swarm.Meta, User]

// creates Golang channels to produce / consume messages
enq, dlq := emit.Event[UserEvent](q)
deq, ack := listen.Event[UserEvent](q)
```

Please see example about event [producer](./broker/sqs/examples/emit/event/sqs.go) and [consumer](./broker/sqs/examples/listen/event/sqs.go).


## Error Handling

The error handling on channel level is governed either by [dead-letter queue](#message-delivery-guarantees) or [acknowledge protocol](#consume-listen-messages). The library provides `swarm.WithStdErr` configuration option to pass the side channel to consume global errors. Use it as top level error handler. 

```go
stderr := make(chan error)
q, err := sqs.New("swarm-test",
  sqs.WithConfig(
    swarm.WithStdErr(stderr),
  ),
)

for err := range stderr {
  // error handling loop
}
```


## Fail Fast

The existing message routing architecture assumes that a micro-batch of messages is read from the broker, dispatched to channels, and then waits for acknowledgments. A new micro-batch is not read until all messages are acknowledged, or the `TimeToFlight` timer expires. In time-critical systems or serverless applications, a "fail fast" strategy is more effective (e.g., a Lambda function doesn't need to idle until the timeout).

Send negative acknowledgement to `ack` channel to indicate error on message processing.

```go
deq, ack := listen.Typed[Note](q)

// consume messages and then acknowledge it
for msg := range deq {
  // negative ack on the error
  if err := doSomething(msg.Object); err != nil {
    ack <- msg.Fail(err)
    continue
  } 
  ack <- msg
}
```


## Serverless 

The library primarily support development of serverless event-driven application using AWS service. The library provides AWS CDK Golang constructs to spawn consumers. See example of [serverless consumer](./broker/eventbridge/examples/listen/typed/eventbridge.go) and corresponding AWS CDK [application](./broker/eventbridge/examples/serverless/eventbridge.go).

It consistently implements a pattern - "create Broker, attach Sinks".  

```go
package main

import (
  "github.com/fogfish/scud"
  "github.com/fogfish/swarm/broker/eventbridge"
)

func main() {
  app := awscdk.NewApp(nil)
  stack := awscdk.NewStack(app, jsii.String("swarm-example-eventbridge"),
    &awscdk.StackProps{
      Env: &awscdk.Environment{
        Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
        Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
      },
    },
 )

  // create broker
  broker := eventbridge.NewBroker(stack, jsii.String("Broker"), nil)
  broker.NewEventBus(nil)

  broker.NewSink(
    &eventbridge.SinkProps{
      Source: []string{"swarm-example-eventbridge"},
      Function: &scud.FunctionGoProps{
        SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
        SourceCodeLambda:  "examples/dequeue/typed",
      },
    },
  )

  app.Synth(nil)
}
```

