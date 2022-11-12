package eventsqs_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/eventsqs"
)

func TestEventBridgeCDK(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := eventsqs.NewServerlessStack(app,
		jsii.String("swarm-example-eventsqs"),
		&eventsqs.ServerlessStackProps{
			Version: "latest",
			System:  "swarm-example-eventsqs",
		},
	)
	stack.NewQueue()

	stack.NewSink(
		&eventsqs.SinkProps{
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/eventsqs/dequeue",
			},
		},
	)

	require := map[*string]*float64{
		jsii.String("AWS::SQS::Queue"):                 jsii.Number(1),
		jsii.String("AWS::Lambda::EventSourceMapping"): jsii.Number(1),
		jsii.String("AWS::IAM::Role"):                  jsii.Number(2),
		jsii.String("AWS::Lambda::Function"):           jsii.Number(2),
		jsii.String("Custom::LogRetention"):            jsii.Number(1),
	}

	template := assertions.Template_FromStack(stack.Stack, nil)
	for key, val := range require {
		template.ResourceCountIs(key, val)
	}
}
