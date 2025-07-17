//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs

import (
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/fogfish/scud"
	"github.com/fogfish/swarm"
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
	Agent    *string
	Queue    awssqs.IQueue
	Function scud.FunctionProps
}

func (props *SinkProps) assert() {
	if props.Function == nil {
		panic("Function is not defined.")
	}
	if props.Queue == nil {
		panic("Queue is not defined.")
	}
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	props.assert()

	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)
	sink.Handler.AddEnvironment(
		jsii.String(EnvConfigSourceSQS),
		props.Queue.QueueName(),
		nil,
	)

	if props.Agent != nil {
		sink.Handler.AddEnvironment(
			jsii.String(swarm.EnvConfigAgent),
			props.Agent,
			nil,
		)
	}

	if ttf := timeToFlight(props.Function); ttf != nil {
		tsf := ttf.ToSeconds(nil)
		tsi := int(aws.ToFloat64(tsf))

		sink.Handler.AddEnvironment(
			jsii.String(swarm.EnvConfigTimeToFlight),
			jsii.String(strconv.Itoa(tsi)),
			nil,
		)
	}

	source := awslambdaeventsources.NewSqsEventSource(props.Queue,
		&awslambdaeventsources.SqsEventSourceProps{})

	sink.Handler.AddEventSource(source)

	return sink
}

func timeToFlight(props scud.FunctionProps) awscdk.Duration {
	if props == nil {
		return nil
	}

	switch v := props.(type) {
	case *scud.FunctionGoProps:
		if v.FunctionProps != nil && v.Timeout != nil {
			return v.Timeout
		}
	case *scud.ContainerGoProps:
		if v.DockerImageFunctionProps != nil && v.Timeout != nil {
			return v.Timeout
		}
	}

	return nil
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

func (broker *Broker) GrantWrite(f awslambda.Function) {
	broker.Queue.GrantSendMessages(f)

	f.AddEnvironment(
		jsii.String(EnvConfigTargetSQS),
		broker.Queue.QueueName(),
		nil,
	)
}

func (broker *Broker) GrantRead(f awslambda.Function) {
	broker.Queue.GrantConsumeMessages(f)

	f.AddEnvironment(
		jsii.String(EnvConfigSourceSQS),
		broker.Queue.QueueName(),
		nil,
	)
}
