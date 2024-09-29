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
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

// Create enqueue routine to WebSocket (AWS API Gateway)
func NewEnqueuer(endpoint string, opts ...Option) (*kernel.Enqueuer, error) {
	cli, err := newWebSocket(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	return kernel.NewEnqueuer(cli, cli.config), nil
}

// Creates dequeue routine from WebSocket (AWS API Gateway)
func NewDequeuer(opts ...Option) (*kernel.Dequeuer, error) {
	c := &Client{}

	for _, opt := range defs {
		opt(c)
	}
	for _, opt := range opts {
		opt(c)
	}

	bridge := &bridge{kernel.NewBridge(c.config.TimeToFlight)}

	return kernel.NewDequeuer(bridge, c.config), nil
}

// Create enqueue & dequeue routines to WebSocket (AWS API Gateway)
func New(endpoint string, opts ...Option) (*kernel.Kernel, error) {
	cli, err := newWebSocket(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	bridge := &bridge{kernel.NewBridge(cli.config.TimeToFlight)}

	return kernel.New(
		kernel.NewEnqueuer(cli, cli.config),
		kernel.NewDequeuer(bridge, cli.config),
	), nil
}

func newWebSocket(endpoint string, opts ...Option) (*Client, error) {
	c := &Client{}

	for _, opt := range defs {
		opt(c)
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.service == nil {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, swarm.ErrServiceIO.New(err)
		}
		c.service = apigatewaymanagementapi.NewFromConfig(cfg,
			func(o *apigatewaymanagementapi.Options) {
				if strings.HasPrefix(endpoint, "wss://") {
					endpoint = strings.Replace(endpoint, "wss://", "https://", 1)
				}

				if strings.HasPrefix(endpoint, "ws://") {
					endpoint = strings.Replace(endpoint, "ws://", "http://", 1)
				}

				o.BaseEndpoint = aws.String(endpoint)
			},
		)
	}

	return c, nil
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
		return swarm.ErrEnqueue.New(err)
	}

	return nil
}

//------------------------------------------------------------------------------

type bridge struct{ *kernel.Bridge }

func (s bridge) Run() { lambda.Start(s.run) }

func (s bridge) run(evt events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	bag := make([]swarm.Bag, 1)
	bag[0] = swarm.Bag{
		Category: evt.RequestContext.RouteKey,
		Digest:   evt.RequestContext.ConnectionID,
		Object:   []byte(evt.Body),
	}

	if err := s.Bridge.Dispatch(bag); err != nil {
		// TODO: handle error
		return events.APIGatewayProxyResponse{StatusCode: http.StatusRequestTimeout}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}
