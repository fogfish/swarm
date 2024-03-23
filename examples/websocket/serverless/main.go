//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/websocket"
)

func main() {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("swarm-example-websocket"),
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			},
		},
	)

	broker := websocket.NewBroker(stack, jsii.String("Broker"), nil)
	broker.NewAuthorizerApiKey("test", "test")

	broker.NewGateway(&websocket.WebSocketApiProps{
		Throttle: &awsapigatewayv2.ThrottleSettings{
			BurstLimit: jsii.Number(1.0),
			RateLimit:  jsii.Number(1.0),
		},
	})

	broker.NewSink(
		&websocket.SinkProps{
			Route: "User",
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/websocket/dequeue",
			},
		},
	)

	app.Synth(nil)
}
