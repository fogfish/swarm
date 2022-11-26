//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events3

import (
	"os"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
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
	Bucket      awss3.Bucket
	EventSource *awslambdaeventsources.S3EventSourceProps
	Lambda      *scud.FunctionGoProps
}

/*
NewSink ...
*/
func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	sink.Handler = scud.NewFunctionGo(sink.Construct, jsii.String("Func"), props.Lambda)

	eventsource := &awslambdaeventsources.S3EventSourceProps{
		Events: &[]awss3.EventType{
			awss3.EventType_OBJECT_CREATED,
		},
	}
	if props.EventSource != nil {
		eventsource = props.EventSource
	}

	sink.Handler.AddEventSource(
		awslambdaeventsources.NewS3EventSource(props.Bucket, eventsource),
	)

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
	bucket awss3.Bucket
	sinks  []*Sink
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

func (stack *ServerlessStack) NewBucket(bucketName ...string) awss3.Bucket {
	name := awscdk.Aws_STACK_NAME()
	if len(bucketName) > 0 {
		name = &bucketName[0]
	}

	stack.bucket = awss3.NewBucket(stack.Stack, jsii.String("Bucket"),
		&awss3.BucketProps{
			BucketName: name,
		},
	)

	return stack.bucket
}

func (stack *ServerlessStack) SetBucket(bucket awss3.Bucket) awss3.Bucket {
	stack.bucket = bucket
	return stack.bucket
}

// func (stack *ServerlessStack) AddBucket(bucketName string) awss3.IBucket {
// 	stack.queue = awssqs.Queue_FromQueueAttributes(stack.Stack, jsii.String("Bucket"),
// 		&awssqs.QueueAttributes{
// 			QueueName: jsii.String(queueName),
// 		},
// 	)

// 	return stack.queue
// }

func (stack *ServerlessStack) NewSink(props *SinkProps) *Sink {
	if stack.bucket == nil {
		panic("Bucket is not defined.")
	}

	props.Bucket = stack.bucket

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

func FromContext(app awscdk.App, key string) string {
	val := app.Node().TryGetContext(jsii.String(key))
	switch v := val.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func FromContextVsn(app awscdk.App) string {
	vsn := FromContext(app, "vsn")
	if vsn == "" {
		return "latest"
	}

	return vsn
}
