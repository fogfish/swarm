//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(
		func(evt events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
			return events.APIGatewayProxyResponse{StatusCode: 200}, nil
		},
	)
}
