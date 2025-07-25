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
    <!-- Go Version -->
    <a href="https://golang.org/dl/">
      <img src="https://img.shields.io/badge/go-1.24+-blue.svg" />
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

# From Chaos to Channels

> *"Finally, messaging that feels like Go!"* — Distributed systems, done right.

Writing distributed, event-driven systems in Go today is harder than it should be - a lesson learned from a decade building such systems in Erlang. Vendor APIs are clunky, non-idiomatic, and tightly coupled to specific messaging brokers.

`swarm` makes asynchronous, distributed messaging in Go **idiomatic, testable, and portable** by expressing queueing/event-driven systems through Go channels instead of vendor-specific APIs.

[User Guide](./doc/user-guide.md) |
[Playground](https://go.dev/play/p/B4C9fcNmWYW) |
[Getting started](#getting-started) | 
[Examples](./broker/) |
[Philosophy](./doc/pattern.md)


## Quick Start

The example below shows the simplest way to enqueue and dequeue messages using AWS SQS.

```go
package main

import (
  "github.com/fogfish/swarm"
  "github.com/fogfish/swarm/broker/sqs"
  "github.com/fogfish/swarm/emit"
  "github.com/fogfish/swarm/listen"
)

// Example message
type Order struct {
  ID     string  `json:"id"`
  Amount float64 `json:"amount"`
}

func main() {
  // create broker for AWS SQS
  q := sqs.Endpoint().Build("aws-sqs-queue-name")

  // create Golang channels
  rcv, ack := listen.Typed[Order](q.Listener)
  out := swarm.LogDeadLetters(emit.Typed[Order](q.Emitter))

  // Send messages
  out <- Order{ID: "123", Amount: 100.0}

  // use Golang channels for I/O
  go func() {
    for order := range rcv {
      processOrder(order.Object)
      ack <- order  // acknowledge processing
    }
  }()

  q.Await()
}

func processOrder(order Order) {
  // Your business logic here
}
```
**That's it!** You're now using distributed messaging with native Go channels.
- Try this code in [playground](https://go.dev/play/p/B4C9fcNmWYW) 
- Continue to [Getting Started](#getting-started) for advanced configuration
- See more examples by browsing [/broker/*/examples/](./broker/)
- Check the design pattern [Distributed event-driven Golang channels](./doc/pattern.md) for deep-dive into library philosophy. 

Continue reading to understand the purpose of the library and why it exists.

<!--
### How it Works

tbd.
-->

## What is `swarm`?

`swarm` is a Go library that solves the complexity of distributed, event-driven systems by abstracting external messaging queues (like AWS SQS, AWS EventBridge, RabbitMQ, Kafka, etc.) behind **type-safe Go channels**. 

By aligning with Go's native concurrency model, `swarm` encourages developers to think in terms of **message flows rather than request-response cycles**, which reflects the asynchronous and unreliable reality of distributed systems. This mindset shift leads to more resilient, scalable, and maintainable architectures — where complexity is not hidden, but made explicit and manageable.  

* **Immediate Productivity**: Use Go channel patterns you already know, no need to learn new APIs.
* **Smooth Testing**: Write unit tests with in-memory channels; switch to real systems for integration.
* **Future-Proof**: Swap messaging technologies without changing your business logic.
* **Serverless-Ready**: Designed for both long-running services and ephemeral serverless functions.
* **Scale-Out**: Idiomatic architecture to build distributed topologies and scale-out Golang applications.

Think of `swarm` as net.Conn **for distributed messaging**: a universal, idiomatic interface you can depend on, regardless of the underlying transport.

Below we discussed why `swarm` exists and what problems it solves.


## Why Choose `swarm`?

Traditional messaging libraries have fundamental issues

| Traditional libraries         | With `swarm`                    |
| ----------------------------- | ------------------------------- |
| Different SDK for each broker | One API for all brokers         |
| Complex mocks for testing     | In-memory channels for tests    |
| Vendor lock-in                | Portable across technologies    |
| Sync APIs for async systems   | Native async with Go channels   |
| Manual error handling         | Built-in retries & dead letters |

`swarm` solves these problems by providing a universal, idiomatic interface built on Go's concurrency model — embracing the asynchronous nature of distributed systems, making message flows explicit, and hiding vendor-specific complexity behind familiar Go channels.

See practical scenarios for `swarm` in [Storytelling](#storytelling---why-we-built-swarm).


## Getting started

The library requires **Go 1.24** or later due to usage of [generic alias types](https://go.dev/blog/alias-names#generic-alias-types).

The latest version of the library is available at `main` branch of this repository. All development, including new features and bug fixes, take place on the `main` branch using forking and pull requests as described in contribution guidelines. The stable version is available via Golang modules.

Use `go get` to retrieve the library and add it as dependency to your application.

```bash
go get -u github.com/fogfish/swarm
```

### When to Use `swarm`

**Perfect for:**
- **Event-driven architectures** - Decouple services with async messaging
- **Serverless applications** - Zero-boilerplate event consumers on AWS Lambda
- **Microservices** - Replace fragile HTTP calls with resilient messaging
- **High-throughput systems** - Scale message processing horizontally
- **Multi-cloud applications** - Abstract away broker differences

**You'll love it if you:**
- Want type-safe messaging without vendor lock-in
- Need to unit test message-driven code easily  
- Prefer Go channels over learning broker-specific APIs
- Value clean, testable architecture patterns

## Advanced Usage and Next steps

* Learn [why we built `swarm`](#storytelling---why-we-built-swarm).
* [User guide](./doc/user-guide.md) help you with adopting the library.
* The library supports a variety of brokers out of the box. See the table at the beginning of this document for details. If you need to implement a broker that isn’t supported yet, refer to [Bring Your Own Broker](./doc/bring-your-own-broker.md) for guidance.


## Real-World Success Stories

Here's how `swarm` solves actual production problems that teams face every day.

### E-commerce Order Processing: The Cascading Failure Problem

Imagine you're building an e-commerce platform that processes thousands of orders per minute during Black Friday sales. Your order processing system needs to:
* Validate payment information
* Update inventory levels
* Send confirmation emails
* Update analytics systems
* Trigger fulfillment processes

What are problems with the traditional approach:
* Blocking Operations: Each service call blocks until completion, making the system slow during peak loads
* Cascading Failures: If the email service is down, the entire order process fails, even though the core business logic succeeded
* Poor Scalability: Cannot scale individual components independently

```go
func ProcessOrder(order *Order) error {
  // Synchronous calls create cascading failures
  if err := paymentService.Charge(order.Payment); err != nil {
    return err // Customer sees error immediately
  }
    
  if err := inventoryService.Reserve(order.Items); err != nil {
    return err // Payment charged but inventory failed
  }
    
  if err := emailService.SendConfirmation(order); err != nil {
    return err // Everything fails if email is slow
  }
    
  return nil
}
```

Using the `swarm` library, we benefit from
* Immediate Response: Customers get instant order confirmation
* Fault Tolerance: If email service is down, orders still process successfully
* Independent Scaling: Each service can scale based on its own load patterns

```go
func ProcessOrder(order *Order) error {
  // Immediate response to customer
  if err := validateOrder(order); err != nil {
    return err
  }

  // Async processing via channels
  orderEvents <- OrderCreated{Order: order}
    
  return nil // Customer gets immediate confirmation
}

// Background processing
go func() {
  for event := range orderEvents {
    paymentQueue <- PaymentRequest{Order: event.Order}
    inventoryQueue <- InventoryUpdate{Order: event.Order}
    emailQueue <- EmailNotification{Order: event.Order}
  }
}()
```

### Serverless Event Processing: The SDK Complexity Problem

You're building a serverless application on AWS that needs to process events from multiple sources - SQS queues, S3 bucket notifications, EventBridge events, and DynamoDB streams. Each service has its own SDK with different patterns.

```go
// Different patterns for each service
func handleSQSEvent(event events.SQSEvent) error {
  for _, record := range event.Records {
    // SQS-specific parsing and acknowledgment
    body := record.Body
    receiptHandle := record.ReceiptHandle
    // ... SQS-specific error handling
  }
}

func handleS3Event(event events.S3Event) error {
  for _, record := range event.Records {
    // S3-specific parsing
    bucket := record.S3.Bucket.Name
    key := record.S3.Object.Key
    // ... S3-specific error handling
  }
}
```

Using the `swarm` library, we benefit from
* Unified Interface: Same code patterns work across all AWS services
* Easy Testing: Mock implementations for local development
* Technology Agnostic: Can switch from SQS to EventBridge without code changes

```go
events := eventbridge.Must(eventbridge.Listener().Build())
upload := s3.Must(s3.Listener().Build())
orders := sqs.Must(sqs.Listener().Build())

listen.Typed[UserEvent](events)
listen.Typed[DataUploaded](upload)
listen.Typed[Order](orders)
```

### Microservices Communication: The Synchronous Trap

You have a microservices architecture where the user service, notification service, and analytics service need to communicate. Traditional REST APIs create tight coupling and cascading failures.

```go
// Synchronous REST calls create brittleness
func CreateUser(user *User) error {
  // Save user
  if err := userDB.Save(user); err != nil {
    return err
  }

  // Synchronous calls to other services
  if err := notificationService.SendWelcomeEmail(user); err != nil {
    return err // User creation fails if email fails
  }

  if err := analyticsService.TrackUserCreated(user); err != nil {
    return err // User creation fails if analytics fails
  }

  return nil
}
```

Using the `swarm` library, we benefit from
* Resilient: Core functionality works even if downstream services are down
* Scalable: Services can process at their own pace
* Evolvable: Easy to add new services without changing existing code

```go
// Async communication via channels
func CreateUser(user *User) error {
  // Core business logic
  if err := userDB.Save(user); err != nil {
      return err
  }

  // Async notifications
  userEvents <- UserCreated{User: user}

  return nil // Immediate success
}

// Services consume events independently
go func() {
  for event := range userEvents {
    notificationQueue <- WelcomeEmail{User: event.User}
    analyticsQueue <- UserCreatedEvent{User: event.User}
  }
}()
```

## How To Contribute

The library is [Apache Version 2.0](LICENSE) licensed and accepts contributions via GitHub pull requests:

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

The build and testing process requires [Go](https://golang.org) version 1.24 or later.

**build** and **test** library.

```bash
git clone https://github.com/fogfish/swarm
cd swarm
go test ./...
```

### commit message

The commit message helps us to write a good release note, speed-up review process. The message should address two question what changed and why. The project follows the template defined by chapter [Contributing to a Project](http://git-scm.com/book/ch5-2.html) of Git book.

### bugs

If you experience any issues with the library, please let us know via [GitHub issues](https://github.com/fogfish/swarm/issues). We appreciate detailed and accurate reports that help us to identity and replicate the issue. 

## Community & Support

- **Documentation**: [User Guide](./doc/user-guide.md) • [API Reference](https://pkg.go.dev/github.com/fogfish/swarm)
- **Questions**: [GitHub Discussions](https://github.com/fogfish/swarm/discussions) for Q&A and ideas
- **Issues**: [GitHub Issues](https://github.com/fogfish/swarm/issues) for bug reports and feature requests
- **Examples**: [Live examples](./broker/) for all supported brokers 


## License

[![See LICENSE](https://img.shields.io/github/license/fogfish/swarm.svg?style=for-the-badge)](LICENSE)

