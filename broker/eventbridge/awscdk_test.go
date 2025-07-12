//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
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
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/eventbridge"
)

func TestEventBridgeCDK(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("text"), &awscdk.StackProps{})

	broker := eventbridge.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewEventBus(nil)

	broker.NewSink(
		&eventbridge.SinkProps{
			EventPattern: eventbridge.Like(
				eventbridge.Source("swarm-example-eventbridge"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/typed",
			},
		},
	)

	require := map[*string]*float64{
		jsii.String("AWS::Events::EventBus"): jsii.Number(1),
		jsii.String("AWS::Events::Rule"):     jsii.Number(1),
		jsii.String("AWS::IAM::Role"):        jsii.Number(1),
		jsii.String("AWS::Lambda::Function"): jsii.Number(1),
	}

	template := assertions.Template_FromStack(stack, nil)
	for key, val := range require {
		template.ResourceCountIs(key, val)
	}
}

func TestNewSinkTimeoutEnvVariable(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("test"), &awscdk.StackProps{})

	broker := eventbridge.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewEventBus(nil)

	// Test with timeout specified
	broker.NewSink(
		&eventbridge.SinkProps{
			EventPattern: eventbridge.Like(
				eventbridge.Source("test-source"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/typed",
				FunctionProps: &awslambda.FunctionProps{
					Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
				},
			},
		},
	)

	template := assertions.Template_FromStack(stack, nil)

	// Verify the timeout environment variable is set correctly
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
		"Environment": map[string]any{
			"Variables": map[string]any{
				swarm.EnvConfigTimeToFlight: "30",
			},
		},
	})
}

func TestAddEventBus(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("test"), &awscdk.StackProps{})

	broker := eventbridge.NewBroker(stack, jsii.String("Test"), nil)

	// Add existing event bus instead of creating new one
	broker.AddEventBus("existing-event-bus")

	broker.NewSink(
		&eventbridge.SinkProps{
			EventPattern: eventbridge.Like(
				eventbridge.Source("test-source"),
			),
			Function: &scud.FunctionGoProps{
				SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
				SourceCodeLambda: "examples/dequeue/typed",
			},
		},
	)

	template := assertions.Template_FromStack(stack, nil)

	// Should have rule but no EventBus resource (since we're using existing one)
	require := map[*string]*float64{
		jsii.String("AWS::Events::EventBus"): jsii.Number(0), // No new event bus created
		jsii.String("AWS::Events::Rule"):     jsii.Number(1), // Rule should still be created
		jsii.String("AWS::Lambda::Function"): jsii.Number(1), // Lambda function + Log Retention function
	}

	for key, val := range require {
		template.ResourceCountIs(key, val)
	}
}

func TestGrantWriteEvents(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("test"), &awscdk.StackProps{})

	broker := eventbridge.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewEventBus(nil)

	// Create a function to grant write permissions to
	fn := scud.NewFunction(stack, jsii.String("TestFunc"), &scud.FunctionGoProps{
		SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
		SourceCodeLambda: "examples/dequeue/typed",
	})

	broker.GrantWriteEvents(fn, "test-agent")

	template := assertions.Template_FromStack(stack, nil)

	// Verify IAM policy is created for write permissions
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]any{
		"PolicyDocument": map[string]any{
			"Statement": []any{
				map[string]any{
					"Action": "events:PutEvents",
					"Effect": "Allow",
				},
			},
		},
	})

	// Verify environment variable is set
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
		"Environment": map[string]any{
			"Variables": map[string]any{
				eventbridge.EnvConfigSourceEventBridge: assertions.Match_AnyValue(),
				eventbridge.EnvConfigEventAgent:        "test-agent",
			},
		},
	})
}

func TestGrantReadEvents(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("test"), &awscdk.StackProps{})

	broker := eventbridge.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewEventBus(nil)

	// Create a function to grant read permissions to
	fn := scud.NewFunction(stack, jsii.String("TestFunc"), &scud.FunctionGoProps{
		SourceCodeModule: "github.com/fogfish/swarm/broker/eventbridge",
		SourceCodeLambda: "examples/dequeue/typed",
	})

	broker.GrantReadEvents(fn, "test-agent")

	template := assertions.Template_FromStack(stack, nil)

	// Verify environment variables are set correctly
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
		"Environment": map[string]any{
			"Variables": map[string]any{
				eventbridge.EnvConfigEventAgent: "test-agent",
			},
		},
	})
}
