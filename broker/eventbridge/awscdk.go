//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
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
	Rule    awsevents.Rule
	Handler awslambda.IFunction
}

// See https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html
type SinkProps struct {
	System       awsevents.IEventBus
	EventPattern *awsevents.EventPattern
	Function     scud.FunctionProps
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	props.Function.Setenv(EnvConfigSourceEventBridge, *props.System.EventBusName())

	pattern := props.EventPattern
	if pattern == nil {
		pattern = &awsevents.EventPattern{
			Account: jsii.Strings(*awscdk.Aws_ACCOUNT_ID()),
		}
	}

	//
	sink.Rule = awsevents.NewRule(sink.Construct, jsii.String("Rule"),
		&awsevents.RuleProps{
			EventBus:     props.System,
			EventPattern: pattern,
		},
	)

	if props.Function != nil {
		sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)

		sink.Rule.AddTarget(awseventstargets.NewLambdaFunction(
			sink.Handler,
			&awseventstargets.LambdaFunctionProps{
				// TODO:
				// MaxEventAge: ,
				// RetryAttempts: ,
			},
		))
	}

	return sink
}

//------------------------------------------------------------------------------
//
// AWS CDK Broker Construct
//
//------------------------------------------------------------------------------

type BrokerProps struct {
	System string
}

type Broker struct {
	constructs.Construct
	Bus awsevents.IEventBus
}

func NewBroker(scope constructs.Construct, id *string, props *BrokerProps) *Broker {
	broker := &Broker{Construct: constructs.NewConstruct(scope, id)}

	return broker
}

func (broker *Broker) NewEventBus(props *awsevents.EventBusProps) awsevents.IEventBus {
	if props == nil {
		props = &awsevents.EventBusProps{}
	}

	if props.EventBusName == nil {
		props.EventBusName = awscdk.Aws_STACK_NAME()
	}

	broker.Bus = awsevents.NewEventBus(broker.Construct, jsii.String("Bus"), props)

	return broker.Bus
}

func (broker *Broker) AddEventBus(eventBusName string) awsevents.IEventBus {
	broker.Bus = awsevents.EventBus_FromEventBusName(broker.Construct, jsii.String("Bus"), jsii.String(eventBusName))

	return broker.Bus
}

func (broker *Broker) NewSink(props *SinkProps) *Sink {
	if broker.Bus == nil {
		panic("EventBus is not defined.")
	}

	props.System = broker.Bus

	name := props.Function.UniqueID()
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}
