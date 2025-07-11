<p align="center">
  <img src="./doc/swarm-v3.png" height="256" />
  <h3 align="center">swarm</h3>
  <p align="center"><strong>Go channels for distributed queueing and event-driven systems</strong></p>

  <p align="center">
    <!-- Build Status  -->
    <a href="https://github.com/fogfish/swarm/actions/">
      <img src="https://github.com/fogfish/swarm/workflows/test/badge.svg" />
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
  </p>

  <table align="center">
    <thead><tr><th>sub-module</th><th>doc</th><th>features</th><th>about</th></tr></thead>
    <tbody>
    <!-- Module swarm + kernel -->
    <tr><td><a href=".">
      <img src="https://img.shields.io/github/v/tag/fogfish/swarm?label=version&filter=v*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/fogfish/swarm">
      <img src="https://img.shields.io/badge/doc-swarm-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td></td>
    <td>
    API & Kernel
    </td></tr>
    <!-- Module broker/eventbridge -->
    <tr><td><a href="./broker/eventbridge/">
      <img src="https://img.shields.io/github/v/tag/fogfish/swarm?label=version&filter=broker/eventbridge/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/eventbridge">
      <img src="https://img.shields.io/badge/doc-eventbridge-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      <img src="https://img.shields.io/badge/rw-9881F3?logo=amazonsqs&logoColor=white&style=platic" />
      <img src="https://img.shields.io/badge/serverless-e999b8?logo=awslambda&logoColor=black&style=platic" />
    </td>
    <td>
    AWS EventBridge 
    </td>
    </tr>
    <!-- Module broker/eventddb -->
    <tr><td><a href="./broker/eventddb/">
      <img src="https://img.shields.io/github/v/tag/fogfish/swarm?label=version&filter=broker/eventddb/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/eventddb">
      <img src="https://img.shields.io/badge/doc-eventddb-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      <img src="https://img.shields.io/badge/ro-4C90F0?logo=amazonsqs&logoColor=white&style=platic" />
      <img src="https://img.shields.io/badge/serverless-e999b8?logo=awslambda&logoColor=black&style=platic" />
    </td>
    <td>
    AWS DynamoDB Stream
    </td></tr>
    <!-- Module broker/events3 -->
    <tr><td><a href="./broker/events3/">
      <img src="https://img.shields.io/github/v/tag/fogfish/swarm?label=version&filter=broker/events3/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/events3">
      <img src="https://img.shields.io/badge/doc-events3-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      <img src="https://img.shields.io/badge/ro-4C90F0?logo=amazonsqs&logoColor=white&style=platic" />
      <img src="https://img.shields.io/badge/serverless-e999b8?logo=awslambda&logoColor=black&style=platic" />
    </td>
    <td>
    AWS S3 Event
    </td></tr>
    <!-- Module broker/eventsqs -->
    <tr><td><a href="./broker/eventsqs/">
      <img src="https://img.shields.io/github/v/tag/fogfish/swarm?label=version&filter=broker/eventsqs/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/eventsqs">
      <img src="https://img.shields.io/badge/doc-eventsqs-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      <img src="https://img.shields.io/badge/ro-4C90F0?logo=amazonsqs&logoColor=white&style=platic" />
      <img src="https://img.shields.io/badge/serverless-e999b8?logo=awslambda&logoColor=black&style=platic" />
    </td>
    <td>
    AWS SQS Events
    </td></tr>
    <!-- Module broker/sqs -->
    <tr><td><a href="./broker/sqs/">
      <img src="https://img.shields.io/github/v/tag/fogfish/swarm?label=version&filter=broker/sqs/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/sqs">
      <img src="https://img.shields.io/badge/doc-sqs-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      <img src="https://img.shields.io/badge/rw-9881F3?logo=amazonsqs&logoColor=white&style=platic" />
    </td>
    <td>
    AWS SQS and SQS FIFO
    </td></tr>
    <!-- Module broker/websocket -->
    <tr><td><a href="./broker/websocket/">
      <img src="https://img.shields.io/github/v/tag/fogfish/swarm?label=version&filter=broker/websocket/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/websocket">
      <img src="https://img.shields.io/badge/doc-websocket-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      <img src="https://img.shields.io/badge/ro-4C90F0?logo=amazonsqs&logoColor=white&style=platic" />
      <img src="https://img.shields.io/badge/serverless-e999b8?logo=awslambda&logoColor=black&style=platic" />
    </td>
    <td>
    AWS WebSocket API
    </td></tr>
    <!-- Module broker/sns -->
    <tr><td><img src="https://img.shields.io/badge/coming%20soon-00b150?style=platic"></td>
    <td><!-- <a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/websocket">
      <img src="https://img.shields.io/badge/doc-websocket-007d9c?logo=go&logoColor=white&style=platic" />
    </a>--></td>
    <td>
      <img src="https://img.shields.io/badge/wo-13C9BA?logo=amazonsqs&logoColor=white&style=platic" />
    </td>
    <td>
    AWS SNS
    </td></tr>
    <!-- Module broker/eventkinesis -->
    <tr><td><img src="https://img.shields.io/badge/coming%20soon-00b150?style=platic"></td>
    <td><!-- <a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/websocket">
      <img src="https://img.shields.io/badge/doc-websocket-007d9c?logo=go&logoColor=white&style=platic" />
    </a>--></td>
    <td>
      <img src="https://img.shields.io/badge/ro-4C90F0?logo=amazonsqs&logoColor=white&style=platic" />
      <img src="https://img.shields.io/badge/serverless-e999b8?logo=awslambda&logoColor=black&style=platic" />
    </td>
    <td>
    AWS Kinesis Events
    </td></tr>
    <!-- Module broker/kinesis -->
    <tr><td><img src="https://img.shields.io/badge/coming%20soon-00b150?style=platic"></td>
    <td><!-- <a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/websocket">
      <img src="https://img.shields.io/badge/doc-websocket-007d9c?logo=go&logoColor=white&style=platic" />
    </a>--></td>
    <td>
      <img src="https://img.shields.io/badge/rw-9881F3?logo=amazonsqs&logoColor=white&style=platic" />
    </td>
    <td>
    AWS Kinesis
    </td></tr>
    <!-- Module broker/elasticache -->
    <tr><td><img src="https://img.shields.io/badge/help%20needed-035392?style=platic"></td>
    <td><!-- <a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/websocket">
      <img src="https://img.shields.io/badge/doc-websocket-007d9c?logo=go&logoColor=white&style=platic" />
    </a>--></td>
    <td>
      <img src="https://img.shields.io/badge/rw-9881F3?logo=amazonsqs&logoColor=white&style=platic" />
    </td>
    <td>
    AWS ElastiCache
    </td></tr>
    <!-- Module broker/mqtt -->
    <tr><td><img src="https://img.shields.io/badge/help%20needed-035392?style=platic"></td>
    <td><!-- <a href="https://pkg.go.dev/github.com/fogfish/swarm/broker/websocket">
      <img src="https://img.shields.io/badge/doc-websocket-007d9c?logo=go&logoColor=white&style=platic" />
    </a>--></td>
    <td>
      <img src="https://img.shields.io/badge/rw-9881F3?logo=amazonsqs&logoColor=white&style=platic" />
    </td>
    <td>
    MQTT
    </td></tr>
    </tbody>
  </table>
</p>

---

Today's wrong abstractions lead to complexity on maintainability in the future. Usage of synchronous interfaces to reflect asynchronous nature of messaging queues is a good example of inaccurate abstraction. Usage of pure Go channels is a proper solution to distills asynchronous semantic of queueing systems into the idiomatic native Golang code. The library adapts Go channels for various systems and interface. Please let us know via [GitHub issues](https://github.com/fogfish/swarm/issue) your needs about queuing technologies.


## Inspiration

The library encourages developers to use Golang struct for asynchronous communication with peers. It helps engineers to define domain models, write correct, maintainable code. This library (`swarm`) uses generic programming style to abstract queueing systems into the idiomatic Golang channels `chan<- T` and `<-chan T`. See the design pattern [Golang channels for distributed event-driven architecture](./doc/pattern.md) to learn philosophy and use-cases:
1. **readability**: application uses pure Go code instead of vendor specific interfaces (learning time) 
2. **portability**: application is portable between various queuing systems or event brokers in same manner as sockets abstracts networking stacks (exchange queueing transport "on-the-fly" to resolve evolution of requirements)  
3. **testability**: unit testing focuses on pure biz logic, simplify dependency injections and mocking (pure unit tests).  
4. **distribution**: idiomatic architecture to build distributed topologies and scale-out Golang applications (clustering).
5. **serverless**: of-the-shelf portable patterns for serverless applications (infrastructure as a code, aws cdk).

## Getting started

The library requires **Go 1.24** or later due to usage of [generic alias types](https://go.dev/blog/alias-names#generic-alias-types).

The latest version of the library is available at `main` branch of this repository. All development, including new features and bug fixes, take place on the `main` branch using forking and pull requests as described in contribution guidelines. The stable version is available via Golang modules.

Use `go get` to retrieve the library and add it as dependency to your application.

```bash
go get -u github.com/fogfish/swarm
```

- [Inspiration](#inspiration)
- [Getting started](#getting-started)
  - [Quick example](#quick-example)
  - [Produce (emit) messages](#produce-emit-messages)
  - [Consume (listen) messages](#consume-listen-messages)
  - [Configure library behavior](#configure-library-behavior)
  - [Message Delivery Guarantees](#message-delivery-guarantees)
  - [Delayed Guarantee vs Guarantee](#delayed-guarantee-vs-guarantee)
  - [Order of Messages](#order-of-messages)
  - [Octet Streams](#octet-streams)
  - [Generic events](#generic-events)
  - [Error Handling](#error-handling)
  - [Fail Fast](#fail-fast)
  - [Serverless](#serverless)
  - [Race condition in Serverless](#race-condition-in-serverless)
- [How To Contribute](#how-to-contribute)
  - [commit message](#commit-message)
  - [bugs](#bugs)
- [Bring Your Own Queue](#bring-your-own-queue)
- [License](#license)

### Quick example

Example below is most simplest illustration of enqueuing and dequeuing message
from AWS SQS.

```go
package main

import (
  "log/slog"

  "github.com/fogfish/swarm"
  "github.com/fogfish/swarm/broker/sqs"
  "github.com/fogfish/swarm/emit"
  "github.com/fogfish/swarm/listen"
)

func main() {
  // create broker for AWS SQS
  q, err := sqs.New("aws-sqs-queue-name")
	if err != nil {
		slog.Error("sqs broker has failed", "err", err)
		return
	}

  // create Golang channels
  rcv, ack := listen.Typed[string](q)
  out := swarm.LogDeadLetters(emit.Typed[string](q))

  // use Golang channels for I/O
  go func() {
    for msg := range rcv {
      out <- msg.Object
      ack <- msg
    }
  }()

  q.Await()
}
```

Check the design pattern [Distributed event-driven Golang channels](./doc/pattern.md) for deep-dive into library philosophy. Also note, each supported broker comes with runnable examples that shows the library. 


### Produce (emit) messages

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
q, err := sqs.New("name-of-the-queue"), /* config options */)

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

### Consume (listen) messages

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
q, err := sqs.New("name-of-the-queue", /* config options */)

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

### Configure library behavior

The library uses "option pattern" for the configuration, which is divided into two parts: a generic I/O kernel configuration and a broker-specific configuration. Please note that each configuration option is prefixed with `With` and implemented in [config.go](./config.go) files.


```go
q, err := sqs.New("name-of-the-queue",
  // WithXXX performs broker configuration
  sqs.WithBatchSize(5),
  // WithConfig performs generic kernel configuration
  sqs.WithConfig(
    swarm.WithSource("name-of-my-component"),
    swarm.WithRetryConstant(10 * time.Millisecond, 3),
    swarm.WithPollFrequency(10 * time.Second),
    /* ... */
  ),
)
```

### Message Delivery Guarantees

Usage of Golang channels as an abstraction raises a concern about grade of service on the message delivery guarantees. The library ensures exactly same grade of service as the underlying queueing system or event broker. Messages are delivered according to the promise once they are accepted by the remote side of queuing system. The library's built-in retry logic protects losses from temporary unavailability of the remote peer. However, Golang channels function as sophisticated "in-memory buffers," which can introduce a delay of a few milliseconds between scheduling a message to the channel and dispatching it to the remote peer. To handle catastrophic failures, choose one of the following policies to either accept or safeguard in-flight messages from potential loss.

**At Most Once** is best effort policy, where a message is published without any formal acknowledgement of receipt, and it isn't replayed. Some messages can be lost as subscribers are not required to acknowledge receipt.  

The library implements asymmetric approaches for message handling. In the **emit** path, buffered Golang channels are used for both message emission and managing dead-letter queues. Similarly, the **listen** path uses buffered Golang channels to deliver messages to the consumer.   

```go
// Spawn a new instance of the messaging broker using At Most Once policy.
// The policy defines the capacity of Golang channel.
q, err := sqs.New("name-of-the-queue",
  swarm.WithPolicyAtMostOnce(1000),
)

// for compatibility reasons two channels are returned on the emit path but
// dead-letter-queue is nil
enq, dlq := emit.Typed[Note](q)

// for compatibility reasons two channels are returned on the listen path but
// ack channel acts as /dev/null discards any sent message
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
q, err := sqs.New("name-of-the-queue",
  swarm.WithPolicyAtLeastOnce(1000),
)

// both channels are unbuffered
enq, dlq := emit.Typed[Note](q)

// buffered channels of capacity n
deq, ack := listen.Typed[Note](q)
```

**Exactly Once** is not supported by the library yet.


### Delayed Guarantee vs Guarantee

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

### Order of Messages

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


### Octet Streams

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


### Generic events

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

// creates Golang channels to produce / consume messages
enq, dlq := emit.Event[Meta, User](q)
deq, ack := listen.Event[Meta, User](q)
```

Please see example about event [producer](./broker/sqs/examples/emit/event/sqs.go) and [consumer](./broker/sqs/examples/listen/event/sqs.go).


### Error Handling

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


### Fail Fast

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


### Serverless 

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



### Race condition in Serverless

In a serverless environment, performing listen and emit operations can lead to race conditions. Specifically, the listen loop may complete before other emitted messages are processed.

```go
rcv, ack := listen.Typed[/* .. */](broker1)
snd, dlq := emit.Typed[/* .. */](broker2)

for msg := range rcv {
  snd <- // ...

  // The ack would cause sleep of function in serverless.
  // snd channel might not be flushed before function sleep.
  // This issue has been resolved with the emission flush mechanism
  // that ensures all pending messages are sent before Lambda suspension.
  ack <- msg   
}
```

**Note**: This race condition has been resolved in the current version. The library now implements an automatic emission flush mechanism in serverless environments that ensures all pending messages are sent to the broker before the Lambda function suspends execution. 



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



## Bring Your Own Queue

TBD

## License

[![See LICENSE](https://img.shields.io/github/license/fogfish/swarm.svg?style=for-the-badge)](LICENSE)

