//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventddb_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/eventddb"
)

func TestEventDdbCDK(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("text"), &awscdk.StackProps{})

	broker := eventddb.NewBroker(stack, jsii.String("Test"), nil)
	broker.NewTable(nil)

	broker.NewSink(
		&eventddb.SinkProps{
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/eventddb/dequeue",
			},
		},
	)

	require := map[*string]*float64{
		jsii.String("AWS::DynamoDB::GlobalTable"):      jsii.Number(1),
		jsii.String("AWS::Lambda::EventSourceMapping"): jsii.Number(1),
		jsii.String("AWS::IAM::Role"):                  jsii.Number(2),
		jsii.String("AWS::Lambda::Function"):           jsii.Number(2),
		jsii.String("Custom::LogRetention"):            jsii.Number(1),
	}

	template := assertions.Template_FromStack(stack, nil)
	for key, val := range require {
		template.ResourceCountIs(key, val)
	}
}
