# Design Pattern: System of Queues

(Message) Queues facilitates async communication in distributed systems (since beginning of compute era).

System design looks on queue as a system capable to "store" and forward message under well-known grade of service (service level objectives).

Async I/O is hard to implement but best reflects properties of real world.

The async I/O pattern does not imply that sender is blocked, follows the status of message on each intermediate node or expects immediate response from recipient. async messages can be lost. the protocol must maintain errors. 

Message Broken is "black" box. The fault-tolerant design requires reliable confirmation about successful scheduling of the message.

Existing Message Queue solution provides "wrong" APIs for developers. It uses sync primitives to reflect async nature of queueing systems.

In Golang, channel is a right abstractions for developers to represent queue.

Sending message is simple as that

```go
queue <- []byte("some message")
```

Receiving message is
```go
for msg := range queue {
  // ...
}
```

anything else is unnecessary complication that misleads developers by forcing usage of sync API primitives. (Hiding async aspects behind vendor api plumbing). Write idiomatic Go code instead of plumbing queue specific APIs

**Retrospective 1** fire and forget

As a developer I want to create queue as a channel so that ...

Imagine SQS API like this

```go
func do(in <-chan []byte) {
  for msg := range in {
    // ...
  }
}


queue := sqs.New(/* ... */)
queue <- []byte("some message")

go do(queue)
```


**Retrospective 2** type-safe 

Single queue is transport consist of multiple topics / channels.
Each channel is bound to well-known kind of message (fixed structure)


**Retrospective 2** fail-safe
