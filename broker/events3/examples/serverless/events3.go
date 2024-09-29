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
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/events3"
)

func main() {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("swarm-example-events3"),
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			},
		},
	)

	broker := events3.NewBroker(stack, jsii.String("Broker"), nil)
	bucket := broker.NewBucket(nil)
	bucket.AddLifecycleRule(&awss3.LifecycleRule{
		Id:         jsii.String("Garbage collector"),
		Enabled:    jsii.Bool(true),
		Expiration: awscdk.Duration_Days(jsii.Number(1.0)),
	})

	broker.NewSink(
		&events3.SinkProps{
			// Note: the default property of EventSource captures OBJECT_CREATED and OBJECT_REMOVED events
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/events3",
				SourceCodeLambda: "examples/dequeue/typed",
			},
		},
	)

	app.Synth(nil)
}
