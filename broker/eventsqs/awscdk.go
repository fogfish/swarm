//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/fogfish/scud"
)

//------------------------------------------------------------------------------
//
// AWS CDK Sink Construct
//
//------------------------------------------------------------------------------

type Sink struct {
	constructs.Construct
	Handler awslambda.Function
}

type SinkProps struct {
	Queue    awssqs.IQueue
	Function scud.FunctionProps
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)
	sink.Handler.AddEnvironment(
		jsii.String(EnvConfigSourceSQS),
		props.Queue.QueueName(),
		nil,
	)

	source := awslambdaeventsources.NewSqsEventSource(props.Queue,
		&awslambdaeventsources.SqsEventSourceProps{})

	sink.Handler.AddEventSource(source)

	return sink
}

//------------------------------------------------------------------------------
//
// AWS CDK Stack Construct
//
//------------------------------------------------------------------------------

type BrokerProps struct{}

type Broker struct {
	constructs.Construct
	Queue awssqs.IQueue
}

func NewBroker(scope constructs.Construct, id *string, props *BrokerProps) *Broker {
	broker := &Broker{Construct: constructs.NewConstruct(scope, id)}

	return broker
}

func (broker *Broker) NewQueue(props *awssqs.QueueProps) awssqs.IQueue {
	if props == nil {
		props = &awssqs.QueueProps{}
	}

	if props.QueueName == nil {
		props.QueueName = awscdk.Aws_STACK_NAME()
	}

	if props.VisibilityTimeout == nil {
		props.VisibilityTimeout = awscdk.Duration_Minutes(jsii.Number(15.0))
	}

	broker.Queue = awssqs.NewQueue(broker.Construct, jsii.String("Queue"), props)

	return broker.Queue
}

func (broker *Broker) AddQueue(queueArn *string) awssqs.IQueue {
	broker.Queue = awssqs.Queue_FromQueueAttributes(broker.Construct, jsii.String("Bus"),
		&awssqs.QueueAttributes{
			QueueArn: queueArn,
		},
	)

	return broker.Queue
}

func (broker *Broker) NewSink(props *SinkProps) *Sink {
	if broker.Queue == nil {
		panic("Queue is not defined.")
	}

	props.Queue = broker.Queue

	name := props.Function.UniqueID()
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}
