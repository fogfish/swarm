//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventsqs"
)

func TestEventBridgeCDK(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("text"), &awscdk.StackProps{})

	broker := eventsqs.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewQueue(nil)

	broker.NewSink(
		&eventsqs.SinkProps{
			Agent: jsii.String("test:agent"),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventsqs",
				SourceCodeLambda: "examples/dequeue/typed",
				FunctionProps: &awslambda.FunctionProps{
					Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
				},
			},
		},
	)

	require := map[*string]*float64{
		jsii.String("AWS::SQS::Queue"):                 jsii.Number(1),
		jsii.String("AWS::Lambda::EventSourceMapping"): jsii.Number(1),
		jsii.String("AWS::IAM::Role"):                  jsii.Number(1),
		jsii.String("AWS::Lambda::Function"):           jsii.Number(1),
	}

	template := assertions.Template_FromStack(stack, nil)
	for key, val := range require {
		template.ResourceCountIs(key, val)
	}

	// Verify the timeout environment variable is set correctly
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
		"Environment": map[string]any{
			"Variables": map[string]any{
				eventsqs.EnvConfigSourceSQS: assertions.Match_ObjectLike(
					&map[string]any{
						"Fn::GetAtt": assertions.Match_AnyValue(),
					},
				),
				swarm.EnvConfigTimeToFlight: "30",
				swarm.EnvConfigAgent:        "test:agent",
			},
		},
	})
}

func TestGrantRead(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("test"), &awscdk.StackProps{})

	broker := eventsqs.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewQueue(nil)

	testFunction := broker.NewSink(
		&eventsqs.SinkProps{
			Agent: jsii.String("test:agent"),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventsqs",
				SourceCodeLambda: "examples/dequeue/typed",
				FunctionProps: &awslambda.FunctionProps{
					Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
				},
			},
		},
	)

	// Grant read permissions
	broker.GrantRead(testFunction.Handler)

	template := assertions.Template_FromStack(stack, nil)

	// Verify environment variable is set
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
		"Environment": map[string]any{
			"Variables": map[string]any{
				eventsqs.EnvConfigSourceSQS: assertions.Match_ObjectLike(
					&map[string]any{
						"Fn::GetAtt": assertions.Match_AnyValue(),
					},
				),
			},
		},
	})
}

func TestGrantWrite(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("test"), &awscdk.StackProps{})

	broker := eventsqs.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewQueue(nil)

	// Create a test Lambda function
	testFunction := broker.NewSink(
		&eventsqs.SinkProps{
			Agent: jsii.String("test:agent"),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventsqs",
				SourceCodeLambda: "examples/dequeue/typed",
				FunctionProps: &awslambda.FunctionProps{
					Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
				},
			},
		},
	)

	// Grant write permissions
	broker.GrantWrite(testFunction.Handler)

	template := assertions.Template_FromStack(stack, nil)

	// Verify environment variable is set
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
		"Environment": map[string]any{
			"Variables": map[string]any{
				eventsqs.EnvConfigTargetSQS: assertions.Match_ObjectLike(
					&map[string]any{
						"Fn::GetAtt": assertions.Match_AnyValue(),
					},
				),
			},
		},
	})
}
