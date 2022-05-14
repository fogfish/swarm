//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs

import (
	"os"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
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
	Queue  awssqs.IQueue
	Lambda *scud.FunctionGoProps
}

/*

NewSink ...
*/
func NewSink(scope constructs.Construct, id *string, props *SinkProps) Sink {
	sink := &sink{Construct: constructs.NewConstruct(scope, id)}

	//
	handler := scud.NewFunctionGo(sink.Construct, jsii.String("Func"), props.Lambda)

	source := awslambdaeventsources.NewSqsEventSource(props.Queue,
		&awslambdaeventsources.SqsEventSourceProps{})

	handler.AddEventSource(source)

	return sink
}

//
type ServerlessApp interface {
	awscdk.App
	CreateQueue() ServerlessApp
	AttachQueue() ServerlessApp
	CreateSink(*SinkProps) ServerlessApp
}

type serverlessapp struct {
	awscdk.App
	stack awscdk.Stack
	sys   string
	queue awssqs.IQueue
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

func (app *serverlessapp) CreateQueue() ServerlessApp {
	app.queue = awssqs.NewQueue(app.stack, jsii.String("Bus"),
		&awssqs.QueueProps{
			QueueName:         awscdk.Aws_STACK_NAME(),
			VisibilityTimeout: awscdk.Duration_Minutes(jsii.Number(15.0)),
		},
	)

	return app
}

func (app *serverlessapp) AttachQueue() ServerlessApp {
	app.queue = awssqs.Queue_FromQueueAttributes(app.stack, jsii.String("Bus"),
		&awssqs.QueueAttributes{
			QueueName: awscdk.Aws_STACK_NAME(),
		},
	)

	return app
}

func (app *serverlessapp) CreateSink(props *SinkProps) ServerlessApp {
	props.Queue = app.queue

	name := "Sink" + strconv.Itoa(len(app.sinks))
	sink := NewSink(app.stack, jsii.String(name), props)

	app.sinks = append(app.sinks, sink)
	return app
}
