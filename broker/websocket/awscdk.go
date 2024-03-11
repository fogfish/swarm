//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package websocket

import (
	"os"
	"strconv"

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
	gateway awsapigatewayv2.WebSocketApi
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
		url := aws.ToString(props.gateway.ApiEndpoint()) + "/" + stage
		(*props.Lambda.FunctionProps.Environment)["CONFIG_SWARM_WS_URL"] = aws.String(url)
	}

	sink.Handler = scud.NewFunctionGo(sink.Construct, jsii.String("Func"), props.Lambda)

	it := integrations.NewWebSocketLambdaIntegration(jsii.String(props.Route), sink.Handler)

	props.gateway.AddRoute(jsii.String(props.Route),
		&awsapigatewayv2.WebSocketRouteOptions{
			Integration: it,
		},
	)

	props.gateway.GrantManageConnections(sink.Handler)

	return sink
}

//------------------------------------------------------------------------------
//
// AWS CDK Stack Construct
//
//------------------------------------------------------------------------------

type ServerlessStackProps struct {
	*awscdk.StackProps
	Version string
	System  string
}

type ServerlessStack struct {
	awscdk.Stack
	acc     int
	Gateway awsapigatewayv2.WebSocketApi
}

func NewServerlessStack(app awscdk.App, id *string, props *ServerlessStackProps) *ServerlessStack {
	sid := *id
	if props.Version != "" {
		sid = sid + "-" + props.Version
	}

	stack := &ServerlessStack{
		Stack: awscdk.NewStack(app, jsii.String(sid), props.StackProps),
	}

	return stack
}

type WebSocketApiProps struct {
	*awsapigatewayv2.WebSocketApiProps
	Throttle   *awsapigatewayv2.ThrottleSettings
	Authorizer string
	Access     string
	Secret     string
}

func (stack *ServerlessStack) NewGateway(props *WebSocketApiProps) awsapigatewayv2.WebSocketApi {
	if props.WebSocketApiProps == nil {
		props.WebSocketApiProps = &awsapigatewayv2.WebSocketApiProps{}
	}

	if props.WebSocketApiProps.ApiName == nil {
		props.ApiName = awscdk.Aws_STACK_NAME()
	}

	if props.WebSocketApiProps.ConnectRouteOptions == nil && props.Authorizer != "" {
		authorizer := scud.NewFunctionGo(stack.Stack, jsii.String("Authorizer"),
			&scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "broker/websocket/lambda/authorizer",
				FunctionProps: &awslambda.FunctionProps{
					Environment: &map[string]*string{
						"CONFIG_SWARM_WS_AUTHORIZER":        jsii.String(props.Authorizer),
						"CONFIG_SWARM_WS_AUTHORIZER_ACCESS": jsii.String(props.Access),
						"CONFIG_SWARM_WS_AUTHORIZER_SECRET": jsii.String(props.Secret),
					},
				},
			},
		)

		connector := scud.NewFunctionGo(stack.Stack, jsii.String("Connector"),
			&scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "broker/websocket/lambda/connector",
			},
		)

		props.WebSocketApiProps.ConnectRouteOptions = &awsapigatewayv2.WebSocketRouteOptions{
			Integration: integrations.NewWebSocketLambdaIntegration(
				jsii.String("defaultConnector"),
				connector,
			),
			Authorizer: authorizers.NewWebSocketLambdaAuthorizer(
				jsii.String("default"),
				authorizer,
				&authorizers.WebSocketLambdaAuthorizerProps{},
			),
		}
	}

	stack.Gateway = awsapigatewayv2.NewWebSocketApi(stack.Stack, jsii.String("Gateway"), props.WebSocketApiProps)

	awsapigatewayv2.NewWebSocketStage(stack.Stack, jsii.String("Stage"),
		&awsapigatewayv2.WebSocketStageProps{
			AutoDeploy:   jsii.Bool(true),
			StageName:    jsii.String(stage),
			Throttle:     props.Throttle,
			WebSocketApi: stack.Gateway,
		},
	)

	return stack.Gateway
}

func (stack *ServerlessStack) NewSink(props *SinkProps) *Sink {
	if stack.Gateway == nil {
		panic("Gatewaye is not defined.")
	}

	props.gateway = stack.Gateway

	stack.acc++
	name := "Sink" + strconv.Itoa(stack.acc)
	sink := NewSink(stack.Stack, jsii.String(name), props)

	return sink
}

//------------------------------------------------------------------------------
//
// AWS CDK App Construct
//
//------------------------------------------------------------------------------

type ServerlessApp struct {
	awscdk.App
}

func NewServerlessApp() *ServerlessApp {
	app := awscdk.NewApp(nil)
	return &ServerlessApp{App: app}
}

func (app *ServerlessApp) NewStack(name string, props ...*awscdk.StackProps) *ServerlessStack {
	config := &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
			Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
		},
	}

	if len(props) == 1 {
		config = props[0]
	}

	return NewServerlessStack(app.App, jsii.String(name), &ServerlessStackProps{
		StackProps: config,
		Version:    FromContextVsn(app),
		System:     name,
	})
}

func FromContext(app awscdk.App, key string) string {
	val := app.Node().TryGetContext(jsii.String(key))
	switch v := val.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func FromContextVsn(app awscdk.App) string {
	vsn := FromContext(app, "vsn")
	if vsn == "" {
		return "latest"
	}

	return vsn
}
