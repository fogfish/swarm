//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/broker/websocket"
)

func main() {
	app := websocket.NewServerlessApp()

	stack := app.NewStack("swarm-example-websocket")
	stack.NewGateway(&websocket.WebSocketApiProps{
		Throttle: &awsapigatewayv2.ThrottleSettings{
			BurstLimit: jsii.Number(1.0),
			RateLimit:  jsii.Number(1.0),
		},
		Authorizer: "Basic",
		Access:     "test",
		Secret:     "test",
	})

	stack.NewSink(
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
