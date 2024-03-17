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
	"log/slog"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// WebSocket declares the subset of interface from AWS SDK used by the lib.
type WebSocket interface {
	PostToConnection(ctx context.Context, params *apigatewaymanagementapi.PostToConnectionInput, optFns ...func(*apigatewaymanagementapi.Options)) (*apigatewaymanagementapi.PostToConnectionOutput, error)
}

type Client struct {
	service WebSocket
	config  swarm.Config
}

func New(endpoint string, opts ...swarm.Option) (swarm.Broker, error) {
	cli, err := NewWebSocket(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	config := swarm.NewConfig()
	for _, opt := range opts {
		opt(&config)
	}

	starter := lambda.Start

	type Mock interface{ Start(interface{}) }
	if config.Service != nil {
		service, ok := config.Service.(Mock)
		if ok {
			starter = service.Start
		}
	}

	sls := spawner{f: starter, c: config}

	return kernel.New(cli, sls, config), err
}

func NewWebSocket(endpoint string, opts ...swarm.Option) (*Client, error) {
	config := swarm.NewConfig()
	for _, opt := range opts {
		opt(&config)
	}

	api, err := newService(endpoint, &config)
	if err != nil {
		return nil, err
	}

	return &Client{
		service: api,
		config:  config,
	}, nil
}

func newService(endpoint string, conf *swarm.Config) (WebSocket, error) {
	if conf.Service != nil {
		service, ok := conf.Service.(WebSocket)
		if ok {
			return service, nil
		}
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, swarm.ErrServiceIO.New(err)
	}

	cli := apigatewaymanagementapi.NewFromConfig(cfg,
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

	return cli, nil
}

// Enq enqueues message to broker
func (cli *Client) Enq(bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout)
	defer cancel()

	slog.Debug("sending ", "cat", bag.Ctx.Category, "obj", bag.Object)

	_, err := cli.service.PostToConnection(ctx,
		&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(bag.Ctx.Category),
			Data:         bag.Object,
		},
	)

	slog.Debug("sending ", "err", err)

	if err != nil {
		return swarm.ErrEnqueue.New(err)
	}

	return nil
}

//------------------------------------------------------------------------------

type WSContext string

const WSRequest = WSContext("WS.Request")

type spawner struct {
	c swarm.Config
	f func(any)
}

func (s spawner) Spawn(k *kernel.Kernel) error {
	s.f(
		func(evt events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
			ctx := swarm.NewContext(
				context.WithValue(context.Background(), WSRequest, evt.RequestContext),
				evt.RequestContext.RouteKey,
				evt.RequestContext.ConnectionID,
			)
			bag := []swarm.Bag{{Ctx: ctx, Object: []byte(evt.Body)}}

			if err := k.Dispatch(bag, s.c.TimeToFlight); err != nil {
				return events.APIGatewayProxyResponse{StatusCode: http.StatusRequestTimeout}, err
			}

			return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
		},
	)

	return nil
}

func (s spawner) Ack(digest string) error   { return nil }
func (s spawner) Ask() ([]swarm.Bag, error) { return nil, nil }
