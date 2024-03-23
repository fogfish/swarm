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
	"github.com/fogfish/swarm/broker/eventsqs"
)

func main() {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("swarm-example-eventsqs"),
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			},
		},
	)

	broker := eventsqs.NewBroker(stack, jsii.String("Broker"), nil)
	broker.NewQueue(nil)

	broker.NewSink(
		&eventsqs.SinkProps{
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/eventsqs/dequeue",
			},
		},
	)

	app.Synth(nil)
}
