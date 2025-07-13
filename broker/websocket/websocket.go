//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package websocket

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// WebSocket declares the subset of interface from AWS SDK used by the lib.
type Gateway interface {
	PostToConnection(ctx context.Context, params *apigatewaymanagementapi.PostToConnectionInput, optFns ...func(*apigatewaymanagementapi.Options)) (*apigatewaymanagementapi.PostToConnectionOutput, error)
}

type Client struct {
	service Gateway
	config  swarm.Config
}

// Enq enqueues message to broker
func (cli *Client) Enq(ctx context.Context, bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(ctx, cli.config.NetworkTimeout)
	defer cancel()

	_, err := cli.service.PostToConnection(ctx,
		&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(bag.Category),
			Data:         bag.Object,
		},
	)

	if err != nil {
		return swarm.ErrEnqueue.With(err)
	}

	return nil
}

//------------------------------------------------------------------------------

type bridge struct{ *kernel.Bridge }

func (s bridge) Run() { lambda.Start(s.run) }

func (s bridge) run(ctx context.Context, evt events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	bag := make([]swarm.Bag, 1)
	bag[0] = swarm.Bag{
		Category:  evt.RequestContext.RouteKey,
		Digest:    evt.RequestContext.ConnectionID,
		IOContext: &evt.RequestContext,
		Object:    []byte(evt.Body),
	}

	if err := s.Bridge.Dispatch(ctx, bag); err != nil {
		// TODO: handle error
		return events.APIGatewayProxyResponse{StatusCode: http.StatusRequestTimeout}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}
