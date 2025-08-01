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

	broker.NewSink(
		&eventbridge.SinkProps{
			Agent: jsii.String("eventbridge:example/typed"),
			EventPattern: eventbridge.Like(
				eventbridge.Category("User", "Note", "Like"),
				eventbridge.Source("eventbridge:example/typed"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/typed",
			},
		},
	)

	broker.NewSink(
		&eventbridge.SinkProps{
			Agent: jsii.String("eventbridge:example/event"),
			EventPattern: eventbridge.Like(
				eventbridge.Category("EventNote"),
				eventbridge.Source("eventbridge:example/event"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/event",
			},
		},
	)

	broker.NewSink(
		&eventbridge.SinkProps{
			Agent: jsii.String("eventbridge:example/bytes"),
			EventPattern: eventbridge.Like(
				eventbridge.Category("User", "Note", "Like"),
				eventbridge.Source("eventbridge:example/bytes"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/bytes",
			},
		},
	)

	app.Synth(nil)
}
