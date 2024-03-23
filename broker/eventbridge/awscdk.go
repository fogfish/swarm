//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
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
	Handler awslambda.IFunction
}

// See https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html
type SinkProps struct {
	System     awsevents.IEventBus
	Source     []string
	Categories []string
	Pattern    map[string]interface{}
	Lambda     *scud.FunctionGoProps
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	//
	sink.Handler = scud.NewFunctionGo(sink.Construct, jsii.String("Func"), props.Lambda)

	//
	pattern := &awsevents.EventPattern{}
	if props.Categories != nil && len(props.Categories) > 0 {
		seq := make([]*string, len(props.Categories))
		for i, category := range props.Categories {
			seq[i] = jsii.String(category)
		}
		pattern.DetailType = &seq
	}

	if props.Source != nil && len(props.Source) > 0 {
		seq := make([]*string, len(props.Source))
		for i, agent := range props.Source {
			seq[i] = jsii.String(agent)
		}
		pattern.Source = &seq
	}

	if props.Pattern != nil {
		pattern.Detail = &props.Pattern
	}

	if pattern.DetailType == nil && pattern.Source == nil {
		pattern.Account = &[]*string{awscdk.Aws_ACCOUNT_ID()}
	}

	//
	rule := awsevents.NewRule(sink.Construct, jsii.String("Rule"),
		&awsevents.RuleProps{
			EventBus:     props.System,
			EventPattern: pattern,
		},
	)
	rule.AddTarget(awseventstargets.NewLambdaFunction(
		sink.Handler,
		&awseventstargets.LambdaFunctionProps{
			// TODO:
			// MaxEventAge: ,
			// RetryAttempts: ,
		},
	))

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
	acc int
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

	broker.acc++
	name := "Sink" + strconv.Itoa(broker.acc)
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}
