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
	Handler awslambda.IFunction
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
func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	sink.Handler = scud.NewFunctionGo(sink.Construct, jsii.String("Func"), props.Lambda)

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

type ServerlessStackProps struct {
	*awscdk.StackProps
	Version string
	System  string
}

type ServerlessStack struct {
	awscdk.Stack
	queue awssqs.IQueue
	sinks []*Sink
}

func NewServerlessStack(app awscdk.App, id *string, props *ServerlessStackProps) *ServerlessStack {
	sid := *id
	if props.Version != "" {
		sid = sid + "-" + props.Version
	}

	stack := &ServerlessStack{
		Stack: awscdk.NewStack(app, jsii.String(sid), props.StackProps),
	}

	return stack
}

func (stack *ServerlessStack) NewQueue(queueName ...string) awssqs.IQueue {
	name := awscdk.Aws_STACK_NAME()
	if len(queueName) > 0 {
		name = &queueName[0]
	}

	stack.queue = awssqs.NewQueue(stack.Stack, jsii.String("Queue"),
		&awssqs.QueueProps{
			QueueName:         name,
			VisibilityTimeout: awscdk.Duration_Minutes(jsii.Number(15.0)),
		},
	)

	return stack.queue
}

func (stack *ServerlessStack) AddQueue(queueName string) awssqs.IQueue {
	stack.queue = awssqs.Queue_FromQueueAttributes(stack.Stack, jsii.String("Bus"),
		&awssqs.QueueAttributes{
			QueueName: jsii.String(queueName),
		},
	)

	return stack.queue
}

func (stack *ServerlessStack) NewSink(props *SinkProps) *Sink {
	if stack.queue == nil {
		panic("Queue is not defined.")
	}

	props.Queue = stack.queue

	name := "Sink" + strconv.Itoa(len(stack.sinks))
	sink := NewSink(stack.Stack, jsii.String(name), props)

	stack.sinks = append(stack.sinks, sink)
	return sink
}

//------------------------------------------------------------------------------
//
// AWS CDK App Construct
//
//------------------------------------------------------------------------------

/*

ServerlessApp ...
*/
type ServerlessApp struct {
	awscdk.App
}

/*

NewServerlessApp ...
*/
func NewServerlessApp() *ServerlessApp {
	app := awscdk.NewApp(nil)
	return &ServerlessApp{App: app}
}

func (app *ServerlessApp) NewStack(name string) *ServerlessStack {
	config := &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
			Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
		},
	}

	return NewServerlessStack(app.App, jsii.String(name), &ServerlessStackProps{
		StackProps: config,
		Version:    FromContextVsn(app),
		System:     name,
	})
}

//
func FromContext(app awscdk.App, key string) string {
	val := app.Node().TryGetContext(jsii.String(key))
	switch v := val.(type) {
	case string:
		return v
	default:
		return ""
	}
}

//
func FromContextVsn(app awscdk.App) string {
	vsn := FromContext(app, "vsn")
	if vsn == "" {
		return "latest"
	}

	return vsn
}

// //
// type ServerlessApp interface {
// 	awscdk.App
// 	CreateQueue() ServerlessApp
// 	AttachQueue(string) ServerlessApp
// 	CreateSink(*SinkProps) ServerlessApp
// }

// type serverlessapp struct {
// 	awscdk.App
// 	stack awscdk.Stack
// 	sys   string
// 	queue awssqs.IQueue
// 	sinks []Sink
// }

// //
// func vsn(app awscdk.App) string {
// 	switch val := app.Node().TryGetContext(jsii.String("vsn")).(type) {
// 	case string:
// 		return val
// 	default:
// 		return "latest"
// 	}
// }

// /*

// NewApp ...
// */
// func NewServerlessApp(sys string) ServerlessApp {
// 	//
// 	// Global config
// 	//
// 	app := awscdk.NewApp(nil)
// 	config := &awscdk.StackProps{
// 		Env: &awscdk.Environment{
// 			Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
// 			Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
// 		},
// 	}

// 	//
// 	// Stack
// 	//
// 	stack := awscdk.NewStack(app, jsii.String(sys+"-"+vsn(app)), config)

// 	return &serverlessapp{
// 		App:   app,
// 		stack: stack,
// 		sys:   sys,
// 		sinks: []Sink{},
// 	}
// }

// func (app *serverlessapp) CreateSink(props *SinkProps) ServerlessApp {
// 	props.Queue = app.queue

// 	name := "Sink" + strconv.Itoa(len(app.sinks))
// 	sink := NewSink(app.stack, jsii.String(name), props)

// 	app.sinks = append(app.sinks, sink)
// 	return app
// }
