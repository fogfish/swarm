//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/events3"
)

func main() {
	app := events3.NewServerlessApp()

	stack := app.NewStack("swarm-example-s3")
	stack.NewBucket()

	stack.NewSink(
		&events3.SinkProps{
			// TODO: make this default
			EventSource: &awslambdaeventsources.S3EventSourceProps{
				Events: &[]awss3.EventType{
					awss3.EventType_OBJECT_CREATED,
				},
			},
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/events3/dequeue",
			},
		},
	)

	app.Synth(nil)
}
