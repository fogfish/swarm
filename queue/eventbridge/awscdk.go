//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"os"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
)

/*

Sink ...
*/
type Sink interface {
	constructs.Construct
}

type sink struct {
	constructs.Construct
}

/*

SinkProps ...
*/
type SinkProps struct {
	System   awsevents.IEventBus
	Queue    string
	Category string
	Lambda   *scud.FunctionGoProps
}

/*

NewSink ...
*/
func NewSink(scope constructs.Construct, id *string, props *SinkProps) Sink {
	sink := &sink{Construct: constructs.NewConstruct(scope, id)}

	//
	handler := scud.NewFunctionGo(sink.Construct, jsii.String("Func"), props.Lambda)

	//
	pattern := &awsevents.EventPattern{}
	if props.Category != "" && props.Category != "*" {
		pattern.DetailType = &[]*string{&props.Category}
	}

	if props.Queue != "" && props.Queue != "*" {
		pattern.Source = &[]*string{&props.Queue}
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
		handler,
		&awseventstargets.LambdaFunctionProps{
			// TODO:
			// MaxEventAge: ,
			// RetryAttempts: ,
		},
	))

	return sink
}

//
type ServerlessApp interface {
	awscdk.App
	CreateEventBus() ServerlessApp
	AttachEventBus() ServerlessApp
	CreateSink(*SinkProps) ServerlessApp
}

type serverlessapp struct {
	awscdk.App
	stack awscdk.Stack
	sys   string
	bus   awsevents.IEventBus
	sinks []Sink
}

/*

NewApp ...
*/
func NewServerlessApp(sys string) ServerlessApp {
	//
	// Global config
	//
	app := awscdk.NewApp(nil)
	config := &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
			Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
		},
	}

	//
	// Stack
	//
	stack := awscdk.NewStack(app, jsii.String(sys), config)

	return &serverlessapp{
		App:   app,
		stack: stack,
		sys:   sys,
		sinks: []Sink{},
	}
}

func (app *serverlessapp) CreateEventBus() ServerlessApp {
	app.bus = awsevents.NewEventBus(app.stack, jsii.String("Bus"),
		&awsevents.EventBusProps{EventBusName: awscdk.Aws_STACK_NAME()},
	)

	return app
}

func (app *serverlessapp) AttachEventBus() ServerlessApp {
	app.bus = awsevents.EventBus_FromEventBusName(app.stack, jsii.String("Bus"),
		awscdk.Aws_STACK_NAME(),
	)

	return app
}

func (app *serverlessapp) CreateSink(props *SinkProps) ServerlessApp {
	props.System = app.bus

	name := "Sink" + strconv.Itoa(len(app.sinks))
	sink := NewSink(app.stack, jsii.String(name), props)

	app.sinks = append(app.sinks, sink)
	return app
}
