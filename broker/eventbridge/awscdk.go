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

/*

SinkProps ...
https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html
*/
type SinkProps struct {
	System     awsevents.IEventBus
	Source     []string
	Categories []string
	Pattern    map[string]interface{}
	Lambda     *scud.FunctionGoProps
}

/*

NewSink ...
*/
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
	bus   awsevents.IEventBus
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

func (stack *ServerlessStack) NewEventBus(eventBusName ...string) awsevents.IEventBus {
	name := awscdk.Aws_STACK_NAME()
	if len(eventBusName) > 0 {
		name = &eventBusName[0]
	}

	stack.bus = awsevents.NewEventBus(stack.Stack, jsii.String("Bus"),
		&awsevents.EventBusProps{EventBusName: name},
	)

	return stack.bus
}

func (stack *ServerlessStack) AddEventBus(eventBusName string) awsevents.IEventBus {
	stack.bus = awsevents.EventBus_FromEventBusName(stack.Stack, jsii.String("Bus"), jsii.String(eventBusName))

	return stack.bus
}

func (stack *ServerlessStack) NewSink(props *SinkProps) *Sink {
	if stack.bus == nil {
		panic("EventBus is not defined.")
	}

	props.System = stack.bus

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
