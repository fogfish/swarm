<p align="center">
  <img src="./doc/swarm-v2.png" height="256" />
  <h3 align="center">swarm</h3>
  <p align="center"><strong>Go channels for queueing systems</strong></p>

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

Today's wrong abstractions lead to complexity on maintainability in the future. Usage of synchronous interfaces to reflect asynchronous nature of messaging queues is a good example of inaccurate abstraction. Usage of pure Go channels is a proper solution to distills asynchronous semantic of queues into the idiomatic native Golang code.

## Inspiration

See [design pattern](./doc/pattern.md) to learn how to improve:

1. **readability**: application uses idiomatic Go code instead of vendor specific interfaces (learning time) 
2. **portability**: application is portable between various queuing systems in same manner as sockets abstracts networking stacks. (exchange tech stack, Develop against an in-memory)
3. **testability**: (unit tests with no dependencies)
4. **serverless** deployments

<!-- TODO: Update inspiration -->

## Getting started

The latest version of the library is available at `main` branch of this repository. All development, including new features and bug fixes, take place on the `main` branch using forking and pull requests as described in contribution guidelines. The stable version is available via Golang modules.

```go
import (
  "github.com/fogfish/swarm"
  "github.com/fogfish/swarm/queue/sqs"
)

// create queueing system and instance of the queue
sys := sqs.NewSystem("swarm-example-sqs")
queue := sqs.Must(sqs.New(sys, "swarm-test"))

// get Go channel to emit messages into queue and receive errors
snd, _ := queue.Send("message-of-type-a")

// get Go channel to recv messages from queue
rcv, ack := queue.Recv("message-of-type-a")

// spawn queue listeners. At this point the system spawns the transport
// routines so that channels are ready for the communications
if err := sys.Listen(); err != nil {
  panic(err)
}

// emit message to the queue
snd <- swarm.Bytes("{\"type\": \"a\", \"some\": \"message\"}")

// receive message from queue
for msg := range rcv {
  // do something with message
  ack <- msg
}

sys.Stop()
```

See [examples](examples) folder for executable examples, code snippets for your projects and receipts to build serverless applications.

## Supported queues

- [x] AWS EventBridge
  - [x] [sending message](examples/eventbridge/send/eventbridge.go)
  - [x] [receiving message](examples/eventbridge/recv/eventbridge.go) using aws lambda
  - [x] [aws cdk construct](examples/eventbridge/serverless/main.go)
- [x] AWS SQS Serverless
  - [x] [sending message](examples/eventsqs/send/eventsqs.go)
  - [x] [receiving message](examples/eventsqs/recv/eventsqs.go) using aws lambda
  - [x] [aws cdk construct](examples/eventsqs/serverless/main.go)
- [x] AWS SQS
  - [x] [sending message](examples/sqs/send/sqs.go)
  - [x] [receiving message](examples/sqs/recv/sqs.go)
- [ ] AWS SNS
  - [ ] sending message
- [ ] AWS Kinesis
  - [ ] sending message
  - [ ] receiving message through lambda handler
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


## How To Abstract Queue

TBD
<!-- TODO -->

## License

[![See LICENSE](https://img.shields.io/github/license/fogfish/swarm.svg?style=for-the-badge)](LICENSE)

