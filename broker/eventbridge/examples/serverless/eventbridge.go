//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/eventbridge"
)

func main() {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("swarm-example-eventbridge"),
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			},
		},
	)

	//
	broker := eventbridge.NewBroker(stack, jsii.String("Broker"), nil)
	broker.NewEventBus(nil)

	typed := broker.NewSink(
		&eventbridge.SinkProps{
			EventPattern: eventbridge.Like(
				eventbridge.Category("User", "Note", "Like"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/typed",
			},
		},
	)
	broker.GrantReadEvents(typed.Handler, "eventbridge:example/typed")

	event := broker.NewSink(
		&eventbridge.SinkProps{
			EventPattern: eventbridge.Like(
				eventbridge.Category("EventNote"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/event",
			},
		},
	)
	broker.GrantReadEvents(event.Handler, "eventbridge:example/event")

	octets := broker.NewSink(
		&eventbridge.SinkProps{
			EventPattern: eventbridge.Like(
				eventbridge.Category("User", "Note", "Like"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/bytes",
			},
		},
	)
	broker.GrantReadEvents(octets.Handler, "eventbridge:example/bytes")

	app.Synth(nil)
}
