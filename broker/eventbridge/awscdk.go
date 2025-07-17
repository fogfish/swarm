//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
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
	Rule    awsevents.Rule
	Handler awslambda.Function
}

// See https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html
type SinkProps struct {
	Agent        *string
	EventBus     awsevents.IEventBus
	EventPattern *awsevents.EventPattern
	Function     scud.FunctionProps
}

func (props *SinkProps) assert() {
	if props.Function == nil {
		panic("Function is not defined.")
	}
	if props.EventBus == nil {
		panic("EventBus is not defined.")
	}
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	props.assert()

	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}
	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)

	sink.Handler.AddEnvironment(
		jsii.String(EnvConfigSourceEventBus),
		props.EventBus.EventBusName(),
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

	pattern := props.EventPattern
	if pattern == nil {
		pattern = &awsevents.EventPattern{
			Account: jsii.Strings(*awscdk.Aws_ACCOUNT_ID()),
		}
	}

	//
	sink.Rule = awsevents.NewRule(sink.Construct, jsii.String("Rule"),
		&awsevents.RuleProps{
			EventBus:     props.EventBus,
			EventPattern: pattern,
		},
	)

	sink.Rule.AddTarget(awseventstargets.NewLambdaFunction(
		sink.Handler,
		&awseventstargets.LambdaFunctionProps{
			// TODO:
			// MaxEventAge: ,
			// RetryAttempts: ,
		},
	))

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
// AWS CDK Broker Construct
//
//------------------------------------------------------------------------------

type BrokerProps struct{}

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

	props.EventBus = broker.Bus

	name := props.Function.UniqueID()
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}

// Grant permission to write events to the EventBus.
func (broker *Broker) GrantWrite(f awslambda.Function) {
	broker.Bus.GrantPutEventsTo(f, nil)

	f.AddEnvironment(
		jsii.String(EnvConfigTargetEventBus),
		broker.Bus.EventBusName(),
		nil,
	)
}

// Grant permission to read events from the EventBus.
func (broker *Broker) GrantRead(f awslambda.Function) {
	f.AddEnvironment(
		jsii.String(EnvConfigSourceEventBus),
		broker.Bus.EventBusName(),
		nil,
	)
}
