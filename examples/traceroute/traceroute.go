//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/broker/eventddb"
	"github.com/fogfish/swarm/broker/events3"
	"github.com/fogfish/swarm/broker/eventsqs"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/broker/websocket"
	"github.com/fogfish/swarm/examples/traceroute/core"
)

var RemovalPolicy = awscdk.RemovalPolicy_DESTROY

const SourceCodeModule = "github.com/fogfish/swarm/examples"

func main() {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("traceroute"),
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			},
			Description: jsii.String("example app for swarm integration testing"),
		},
	)

	NewTraceRoute(stack)

	app.Synth(nil)
}

type TraceRoute struct {
	websocket   *WebSocket
	eventbridge *EventBridge
	eventddb    *EventDDB
	events3     *EventS3
	eventsqs    *EventSQS
	sqs         *SQS
}

func NewTraceRoute(stack awscdk.Stack) *TraceRoute {
	c := &TraceRoute{
		websocket:   NewWebSocket(stack),
		eventbridge: NewEventBridge(stack),
		eventddb:    NewEventDDB(stack),
		events3:     NewEventS3(stack),
		eventsqs:    NewEventSQS(stack),
		sqs:         NewSQS(stack),
	}

	c.allowWriteToWebSocket()
	c.allowWriteToNeigboor()

	return c
}

// TODO: add log group for each lambda function
// func (c *TraceRoute) logGroup(stack awscdk.Stack) {
//
// }

func (c *TraceRoute) allowWriteToWebSocket() {
	for _, f := range []awslambda.Function{
		c.websocket.sink.Handler,
		c.eventbridge.sink.Handler,
		c.eventddb.sink.Handler,
		c.events3.sink.Handler,
		c.eventsqs.sink.Handler,
		c.sqs.sink.Handler,
	} {
		c.websocket.broker.GrantWrite(f)
	}
}

func (c *TraceRoute) allowWriteToNeigboor() {
	c.eventbridge.broker.GrantWrite(c.websocket.sink.Handler)

	c.eventddb.broker.GrantWrite(c.eventbridge.sink.Handler)

	c.events3.broker.GrantWrite(c.eventddb.sink.Handler)

	c.eventsqs.broker.GrantWrite(c.events3.sink.Handler)

	c.sqs.broker.GrantWrite(c.eventsqs.sink.Handler)
}

//------------------------------------------------------------------------------

type WebSocket struct {
	broker *websocket.Broker
	sink   *websocket.Sink
}

func NewWebSocket(stack awscdk.Stack) *WebSocket {
	b := &WebSocket{}
	b.createBroker(stack)
	b.createSink()

	return b
}

func (b *WebSocket) createBroker(stack awscdk.Stack) {
	b.broker = websocket.NewBroker(stack, jsii.String("Gateway"), nil)
	b.broker.NewAuthorizerApiKey(
		&websocket.AuthorizerApiKeyProps{
			Access: "traceroute",
			Secret: os.Getenv("TRACEROUTE_API_KEY"),
			Scope:  []string{"it"},
		},
	)

	b.broker.NewGateway(
		&websocket.WebSocketApiProps{},
	)
}

func (b *WebSocket) createSink() {
	b.sink = b.broker.NewSink(
		&websocket.SinkProps{
			Agent:    jsii.String("tr:websocket"),
			Category: swarm.TypeOf[core.Request](),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: SourceCodeModule,
				SourceCodeLambda: "traceroute/cmd/websocket",
			},
		},
	)
}

//------------------------------------------------------------------------------

type EventBridge struct {
	broker *eventbridge.Broker
	sink   *eventbridge.Sink
}

func NewEventBridge(stack awscdk.Stack) *EventBridge {
	b := &EventBridge{}
	b.createBroker(stack)
	b.createSink()

	return b
}

func (b *EventBridge) createBroker(stack awscdk.Stack) {
	b.broker = eventbridge.NewBroker(stack, jsii.String("EventBridge"), nil)
	b.broker.NewEventBus(nil)
}

func (b *EventBridge) createSink() {
	b.sink = b.broker.NewSink(
		&eventbridge.SinkProps{
			Agent: jsii.String("tr:eventbridge"),
			EventPattern: eventbridge.Like(
				eventbridge.Event[core.ReqTrace](),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: SourceCodeModule,
				SourceCodeLambda: "traceroute/cmd/eventbridge",
			},
		},
	)

	b.broker.GrantRead(b.sink.Handler)
}

//------------------------------------------------------------------------------

type EventDDB struct {
	broker *eventddb.Broker
	sink   *eventddb.Sink
}

func NewEventDDB(stack awscdk.Stack) *EventDDB {
	b := &EventDDB{}
	b.createBroker(stack)
	b.createSink()

	return b
}

func (b *EventDDB) createBroker(stack awscdk.Stack) {
	b.broker = eventddb.NewBroker(stack, jsii.String("Dynamo"), nil)
	b.broker.NewTable(
		&awsdynamodb.TablePropsV2{
			TimeToLiveAttribute: jsii.String("ttl"),
			RemovalPolicy:       RemovalPolicy,
		},
	)
}

func (b *EventDDB) createSink() {
	b.sink = b.broker.NewSink(
		&eventddb.SinkProps{
			Agent: jsii.String("tr:eventddb"),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: SourceCodeModule,
				SourceCodeLambda: "traceroute/cmd/eventddb",
			},
		},
	)
}

//------------------------------------------------------------------------------

type EventS3 struct {
	broker *events3.Broker
	sink   *events3.Sink
}

func NewEventS3(stack awscdk.Stack) *EventS3 {
	b := &EventS3{}
	b.createBroker(stack)
	b.createSink()

	return b
}

func (b *EventS3) createBroker(stack awscdk.Stack) {
	b.broker = events3.NewBroker(stack, jsii.String("S3"), nil)
	b.broker.NewBucket(
		&awss3.BucketProps{
			BucketName:    jsii.String("traceroute-swarm-it"),
			RemovalPolicy: RemovalPolicy,
		},
	)
}

func (b *EventS3) createSink() {
	b.sink = b.broker.NewSink(
		&events3.SinkProps{
			Agent: jsii.String("tr:events3"),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: SourceCodeModule,
				SourceCodeLambda: "traceroute/cmd/events3",
			},
		},
	)
	b.broker.GrantRead(b.sink.Handler)
}

//------------------------------------------------------------------------------

type EventSQS struct {
	broker *eventsqs.Broker
	sink   *eventsqs.Sink
}

func NewEventSQS(stack awscdk.Stack) *EventSQS {
	b := &EventSQS{}
	b.createBroker(stack)
	b.createSink()

	return b
}

func (b *EventSQS) createBroker(stack awscdk.Stack) {
	b.broker = eventsqs.NewBroker(stack, jsii.String("EventSQS"), nil)
	b.broker.NewQueue(
		&awssqs.QueueProps{
			RemovalPolicy: RemovalPolicy,
		},
	)
}

func (b *EventSQS) createSink() {
	b.sink = b.broker.NewSink(
		&eventsqs.SinkProps{
			Agent: jsii.String("tr:eventsqs"),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: SourceCodeModule,
				SourceCodeLambda: "traceroute/cmd/eventsqs",
			},
		},
	)
	b.broker.GrantRead(b.sink.Handler)
}

//------------------------------------------------------------------------------

type SQS struct {
	broker *eventsqs.Broker
	queue  awssqs.Queue
	sink   *eventsqs.Sink
}

func NewSQS(stack awscdk.Stack) *SQS {
	b := &SQS{}
	b.createBroker(stack)
	b.createQueue(stack)
	b.createSink()

	return b
}

func (b *SQS) createBroker(stack awscdk.Stack) {
	b.broker = eventsqs.NewBroker(stack, jsii.String("QueueCtrl"), nil)
	b.broker.NewQueue(
		&awssqs.QueueProps{
			QueueName:     jsii.String("traceroute-ctrl"),
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		},
	)
}

func (b *SQS) createQueue(stack awscdk.Stack) {
	b.queue = awssqs.NewQueue(stack, jsii.String("QueueData"),
		&awssqs.QueueProps{
			QueueName:     jsii.String("traceroute-data"),
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		},
	)
}

func (b *SQS) createSink() {
	b.sink = b.broker.NewSink(
		&eventsqs.SinkProps{
			Agent: jsii.String("tr:sqs"),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: SourceCodeModule,
				SourceCodeLambda: "traceroute/cmd/sqs",
				FunctionProps: &awslambda.FunctionProps{
					Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
				},
			},
		},
	)

	b.broker.GrantRead(b.sink.Handler)
	b.queue.GrantSendMessages(b.sink.Handler)
	b.queue.GrantConsumeMessages(b.sink.Handler)
	b.sink.Handler.AddEnvironment(
		jsii.String(sqs.EnvConfigSourceSQS),
		b.queue.QueueName(),
		nil,
	)
	b.sink.Handler.AddEnvironment(
		jsii.String(sqs.EnvConfigTargetSQS),
		b.queue.QueueName(),
		nil,
	)
}
