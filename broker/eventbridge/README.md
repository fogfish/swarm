# AWS EventBridge Broker

The sub-module implements swarm broker for AWS EventBridge. See [the library documentation](../../README.md) for details.


## Serverless


**AWS Event Bridge** has a feature that allows to [match execution of consumer to the pattern](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/CloudWatchEventsandEventPatterns.html#CloudWatchEventsPatterns) of JSON object. Use it to build reliable matching of incoming events:

```go
/*
enq <- &swarm.Event[User]{
  Meta: &swarm.Meta{Agent: "swarm:example", Participant: "user"},
  Data: &User{ID: "user", Text: "some text"},
}
*/

stack.NewSink(
  &eventbridge.SinkProps{
    EventPattern: eventbridge.Event(
      eventbridge.Type[User](),
      eventbridge.Meta("agent", curie.IRI("swarm:example"))
      eventbridge.Meta("participant", curie.IRI("user"))
    ),
    /* ... */
  },
)
```


## Limitation

AWS EventBridge cannot transmit bytes. They have to be enveloped as JSON object. 
The library implements automatic encapsulation using built-in `kernel/encoding/CodecPacket`. Use these codec (`WithPacketCodec`) if you automatically source events from EventBridge to SQS.
