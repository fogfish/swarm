//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events3_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/events3"
)

func TestEventBridgeCDK(t *testing.T) {
	app := events3.NewServerlessApp()
	stack := app.NewStack("swarm-example-events3", nil)
	stack.NewBucket()

	stack.NewSink(
		&events3.SinkProps{
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/events3/dequeue",
			},
		},
	)

	require := map[*string]*float64{
		jsii.String("AWS::S3::Bucket"):               jsii.Number(1),
		jsii.String("Custom::S3BucketNotifications"): jsii.Number(1),
		jsii.String("AWS::IAM::Role"):                jsii.Number(3),
		jsii.String("AWS::Lambda::Function"):         jsii.Number(3),
		jsii.String("Custom::LogRetention"):          jsii.Number(1),
	}

	template := assertions.Template_FromStack(stack.Stack, nil)
	for key, val := range require {
		template.ResourceCountIs(key, val)
	}
}
