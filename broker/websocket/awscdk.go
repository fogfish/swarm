//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package websocket

import (
	"strconv"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	authorizers "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2authorizers"
	integrations "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
)

//------------------------------------------------------------------------------
//
// AWS CDK Sink Construct
//
//------------------------------------------------------------------------------

const stage = "ws"

type Sink struct {
	constructs.Construct
	Handler awslambda.IFunction
}

type SinkProps struct {
	Route   string
	Lambda  *scud.FunctionGoProps
	Gateway awsapigatewayv2.WebSocketApi
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	if props.Lambda.FunctionProps == nil {
		props.Lambda.FunctionProps = &awslambda.FunctionProps{}
	}

	if props.Lambda.FunctionProps.Environment == nil {
		props.Lambda.FunctionProps.Environment = &map[string]*string{}
	}

	if _, has := (*props.Lambda.FunctionProps.Environment)["CONFIG_SWARM_WS_EVENT_TYPE"]; !has {
		(*props.Lambda.FunctionProps.Environment)["CONFIG_SWARM_WS_EVENT_TYPE"] = jsii.String(props.Route)
	}

	if _, has := (*props.Lambda.FunctionProps.Environment)["CONFIG_SWARM_WS_URL"]; !has {
		url := aws.ToString(props.Gateway.ApiEndpoint()) + "/" + stage
		(*props.Lambda.FunctionProps.Environment)["CONFIG_SWARM_WS_URL"] = aws.String(url)
	}

	sink.Handler = scud.NewFunctionGo(sink.Construct, jsii.String("Func"), props.Lambda)

	it := integrations.NewWebSocketLambdaIntegration(jsii.String(props.Route), sink.Handler)

	props.Gateway.AddRoute(jsii.String(props.Route),
		&awsapigatewayv2.WebSocketRouteOptions{
			Integration: it,
		},
	)

	props.Gateway.GrantManageConnections(sink.Handler)

	return sink
}

//------------------------------------------------------------------------------
//
// AWS CDK Stack Construct
//
//------------------------------------------------------------------------------

type BrokerProps struct {
	System string
}

type Broker struct {
	constructs.Construct
	Gateway    awsapigatewayv2.WebSocketApi
	Authorizer awsapigatewayv2.IWebSocketRouteAuthorizer
	acc        int
}

func NewBroker(scope constructs.Construct, id *string, props *BrokerProps) *Broker {
	broker := &Broker{Construct: constructs.NewConstruct(scope, id)}

	return broker
}

func (broker *Broker) NewAuthorizerApiKey(access, secret string) awsapigatewayv2.IWebSocketRouteAuthorizer {
	if broker.Gateway != nil {
		panic("Authorizer MUST be defined before the gateway is instantiated.")
	}

	handler := scud.NewFunctionGo(broker.Construct, jsii.String("Authorizer"),
		&scud.FunctionGoProps{
			SourceCodePackage: "github.com/fogfish/swarm",
			SourceCodeLambda:  "broker/websocket/lambda/authkey",
			FunctionProps: &awslambda.FunctionProps{
				Environment: &map[string]*string{
					"CONFIG_SWARM_WS_AUTHORIZER_ACCESS": jsii.String(secret),
					"CONFIG_SWARM_WS_AUTHORIZER_SECRET": jsii.String(secret),
				},
			},
		},
	)

	broker.Authorizer = authorizers.NewWebSocketLambdaAuthorizer(
		jsii.String("default"),
		handler,
		&authorizers.WebSocketLambdaAuthorizerProps{
			IdentitySource: jsii.Strings("route.request.querystring.apikey"),
		},
	)

	return broker.Authorizer
}

func (broker *Broker) NewAuthorizerJWT(issuer, audience string) awsapigatewayv2.IWebSocketRouteAuthorizer {
	if broker.Gateway != nil {
		panic("Authorizer MUST be defined before the gateway is instantiated.")
	}

	if !strings.HasPrefix(issuer, "https://") {
		panic("Issuer URL MUST start with https://")
	}

	if !strings.HasSuffix(issuer, "/") {
		issuer += "/"
	}

	handler := scud.NewFunctionGo(broker.Construct, jsii.String("Authorizer"),
		&scud.FunctionGoProps{
			SourceCodePackage: "github.com/fogfish/swarm",
			SourceCodeLambda:  "broker/websocket/lambda/authjwt",
			FunctionProps: &awslambda.FunctionProps{
				Environment: &map[string]*string{
					"CONFIG_SWARM_WS_AUTHORIZER_ISS": jsii.String(issuer),
					"CONFIG_SWARM_WS_AUTHORIZER_AUD": jsii.String(audience),
				},
			},
		},
	)

	broker.Authorizer = authorizers.NewWebSocketLambdaAuthorizer(
		jsii.String("default"),
		handler,
		&authorizers.WebSocketLambdaAuthorizerProps{
			IdentitySource: jsii.Strings("route.request.querystring.token"),
		},
	)

	return broker.Authorizer
}

type WebSocketApiProps struct {
	*awsapigatewayv2.WebSocketApiProps
	Throttle *awsapigatewayv2.ThrottleSettings
}

func (broker *Broker) NewGateway(props *WebSocketApiProps) awsapigatewayv2.WebSocketApi {
	if props.WebSocketApiProps == nil {
		props.WebSocketApiProps = &awsapigatewayv2.WebSocketApiProps{}
	}

	if props.WebSocketApiProps.ApiName == nil {
		props.ApiName = awscdk.Aws_STACK_NAME()
	}

	if props.WebSocketApiProps.ConnectRouteOptions == nil && broker.Authorizer != nil {
		connector := scud.NewFunctionGo(broker.Construct, jsii.String("Connector"),
			&scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "broker/websocket/lambda/connector",
			},
		)

		props.WebSocketApiProps.ConnectRouteOptions = &awsapigatewayv2.WebSocketRouteOptions{
			Integration: integrations.NewWebSocketLambdaIntegration(jsii.String("defcon"), connector),
			Authorizer:  broker.Authorizer,
		}
	}

	broker.Gateway = awsapigatewayv2.NewWebSocketApi(broker.Construct, jsii.String("Gateway"), props.WebSocketApiProps)

	awsapigatewayv2.NewWebSocketStage(broker.Construct, jsii.String("Stage"),
		&awsapigatewayv2.WebSocketStageProps{
			AutoDeploy:   jsii.Bool(true),
			StageName:    jsii.String(stage),
			Throttle:     props.Throttle,
			WebSocketApi: broker.Gateway,
		},
	)

	return broker.Gateway
}

func (broker *Broker) NewSink(props *SinkProps) *Sink {
	if broker.Gateway == nil {
		panic("Gatewaye is not defined.")
	}

	props.Gateway = broker.Gateway

	broker.acc++
	name := "Sink" + strconv.Itoa(broker.acc)
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}
