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

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/events3"
)

func TestEventS3CDK(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("text"), &awscdk.StackProps{})

	broker := events3.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewBucket(nil)

	broker.NewSink(
		&events3.SinkProps{
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/events3",
				SourceCodeLambda: "examples/dequeue/typed",
			},
		},
	)

	require := map[*string]*float64{
		jsii.String("AWS::S3::Bucket"):               jsii.Number(1),
		jsii.String("Custom::S3BucketNotifications"): jsii.Number(1),
		jsii.String("AWS::IAM::Role"):                jsii.Number(2),
		jsii.String("AWS::Lambda::Function"):         jsii.Number(2),
	}

	template := assertions.Template_FromStack(stack, nil)
	for key, val := range require {
		template.ResourceCountIs(key, val)
	}
}
