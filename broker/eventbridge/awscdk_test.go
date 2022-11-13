//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/eventbridge"
)

func TestEventBridgeCDK(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := eventbridge.NewServerlessStack(app,
		jsii.String("swarm-example-eventbridge"),
		&eventbridge.ServerlessStackProps{
			Version: "latest",
			System:  "swarm-example-eventbridge",
		},
	)
	stack.NewEventBus()

	stack.NewSink(
		&eventbridge.SinkProps{
			Source: []string{"swarm-example-eventbridge"},
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/eventbridge/dequeue",
			},
		},
	)

	require := map[*string]*float64{
		jsii.String("AWS::Events::EventBus"): jsii.Number(1),
		jsii.String("AWS::Events::Rule"):     jsii.Number(1),
		jsii.String("AWS::IAM::Role"):        jsii.Number(2),
		jsii.String("AWS::Lambda::Function"): jsii.Number(2),
		jsii.String("Custom::LogRetention"):  jsii.Number(1),
	}

	template := assertions.Template_FromStack(stack.Stack, nil)
	for key, val := range require {
		template.ResourceCountIs(key, val)
	}
}