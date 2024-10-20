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
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
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
	Route    string
	Function scud.FunctionProps
	Gateway  awsapigatewayv2.WebSocketApi
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	props.Function.Setenv(EnvConfigEventType, props.Route)
	props.Function.Setenv(EnvConfigSourceWebSocket, aws.ToString(props.Gateway.ApiEndpoint())+"/"+stage)

	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)

	it := integrations.NewWebSocketLambdaIntegration(jsii.String(props.Route), sink.Handler,
		&integrations.WebSocketLambdaIntegrationProps{},
	)

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
	domain     awsapigatewayv2.DomainName
	dns        awsroute53.ARecord
	acc        int
}

func NewBroker(scope constructs.Construct, id *string, props *BrokerProps) *Broker {
	broker := &Broker{Construct: constructs.NewConstruct(scope, id)}

	return broker
}

type AuthorizerApiKeyProps struct {
	Access string
	Secret string
	Scope  []string
}

func (broker *Broker) NewAuthorizerApiKey(props *AuthorizerApiKeyProps) awsapigatewayv2.IWebSocketRouteAuthorizer {
	if broker.Gateway != nil {
		panic("Authorizer MUST be defined before the gateway is instantiated.")
	}

	if props.Access == "" || props.Secret == "" {
		panic("Authorizer MUST define access and secret api keys")
	}

	handler := scud.NewFunctionGo(broker.Construct, jsii.String("Authorizer"),
		&scud.FunctionGoProps{
			SourceCodeModule: "github.com/fogfish/swarm",
			SourceCodeLambda: "broker/websocket/lambda/auth",
			FunctionProps: &awslambda.FunctionProps{
				Environment: &map[string]*string{
					"CONFIG_SWARM_WS_AUTHORIZER_ACCESS": jsii.String(props.Access),
					"CONFIG_SWARM_WS_AUTHORIZER_SECRET": jsii.String(props.Secret),
					"CONFIG_SWARM_WS_AUTHORIZER_SCOPE":  jsii.String(strings.Join(props.Scope, " ")),
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

type AuthorizerJwtProps struct {
	Issuer   string
	Audience string
}

func (broker *Broker) NewAuthorizerJwt(props *AuthorizerJwtProps) awsapigatewayv2.IWebSocketRouteAuthorizer {
	if broker.Gateway != nil {
		panic("Authorizer MUST be defined before the gateway is instantiated.")
	}

	if !strings.HasPrefix(props.Issuer, "https://") {
		panic("Issuer URL MUST start with https://")
	}

	if !strings.HasSuffix(props.Issuer, "/") {
		props.Issuer += "/"
	}

	handler := scud.NewFunctionGo(broker.Construct, jsii.String("Authorizer"),
		&scud.FunctionGoProps{
			SourceCodeModule: "github.com/fogfish/swarm",
			SourceCodeLambda: "broker/websocket/lambda/auth",
			FunctionProps: &awslambda.FunctionProps{
				Environment: &map[string]*string{
					"CONFIG_SWARM_WS_AUTHORIZER_ISS": jsii.String(props.Issuer),
					"CONFIG_SWARM_WS_AUTHORIZER_AUD": jsii.String(props.Audience),
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

type AuthorizerUniversalProps struct {
	AuthorizerApiKey *AuthorizerApiKeyProps
	AuthorizerJwt    *AuthorizerJwtProps
}

func (broker *Broker) NewAuthorizerUniversal(props *AuthorizerUniversalProps) awsapigatewayv2.IWebSocketRouteAuthorizer {
	if broker.Gateway != nil {
		panic("Authorizer MUST be defined before the gateway is instantiated.")
	}

	if props.AuthorizerApiKey == nil || props.AuthorizerJwt == nil {
		panic("Universal Authorizer requires definition of all members")
	}

	if props.AuthorizerApiKey.Access == "" || props.AuthorizerApiKey.Secret == "" {
		panic("Authorizer MUST define access and secret api keys")
	}

	if !strings.HasPrefix(props.AuthorizerJwt.Issuer, "https://") {
		panic("Issuer URL MUST start with https://")
	}

	if !strings.HasSuffix(props.AuthorizerJwt.Issuer, "/") {
		props.AuthorizerJwt.Issuer += "/"
	}

	handler := scud.NewFunctionGo(broker.Construct, jsii.String("Authorizer"),
		&scud.FunctionGoProps{
			SourceCodeModule: "github.com/fogfish/swarm",
			SourceCodeLambda: "broker/websocket/lambda/auth",
			FunctionProps: &awslambda.FunctionProps{
				Environment: &map[string]*string{
					"CONFIG_SWARM_WS_AUTHORIZER_ACCESS": jsii.String(props.AuthorizerApiKey.Access),
					"CONFIG_SWARM_WS_AUTHORIZER_SECRET": jsii.String(props.AuthorizerApiKey.Secret),
					"CONFIG_SWARM_WS_AUTHORIZER_ISS":    jsii.String(props.AuthorizerJwt.Issuer),
					"CONFIG_SWARM_WS_AUTHORIZER_AUD":    jsii.String(props.AuthorizerJwt.Audience),
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
	Host     *string
	TlsArn   *string
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
				SourceCodeModule: "github.com/fogfish/swarm",
				SourceCodeLambda: "broker/websocket/lambda/connector",
			},
		)

		props.WebSocketApiProps.ConnectRouteOptions = &awsapigatewayv2.WebSocketRouteOptions{
			Integration: integrations.NewWebSocketLambdaIntegration(jsii.String("defcon"), connector,
				&integrations.WebSocketLambdaIntegrationProps{},
			),
			Authorizer: broker.Authorizer,
		}
	}

	broker.Gateway = awsapigatewayv2.NewWebSocketApi(broker.Construct, jsii.String("Gateway"), props.WebSocketApiProps)

	var domain *awsapigatewayv2.DomainMappingOptions
	if props.Host != nil && props.TlsArn != nil {
		broker.domain = awsapigatewayv2.NewDomainName(broker.Construct, jsii.String("DomainName"),
			&awsapigatewayv2.DomainNameProps{
				EndpointType: awsapigatewayv2.EndpointType_REGIONAL,
				DomainName:   props.Host,
				Certificate:  awscertificatemanager.Certificate_FromCertificateArn(broker.Construct, jsii.String("X509"), props.TlsArn),
			},
		)

		domain = &awsapigatewayv2.DomainMappingOptions{
			DomainName: broker.domain,
		}
	}

	awsapigatewayv2.NewWebSocketStage(broker.Construct, jsii.String("Stage"),
		&awsapigatewayv2.WebSocketStageProps{
			AutoDeploy:    jsii.Bool(true),
			StageName:     jsii.String(stage),
			Throttle:      props.Throttle,
			WebSocketApi:  broker.Gateway,
			DomainMapping: domain,
		},
	)

	if props.Host != nil && props.TlsArn != nil {
		broker.createRoute53(*props.Host)
	}

	return broker.Gateway
}

func (broker *Broker) createRoute53(host string) {
	domain := strings.Join(strings.Split(host, ".")[1:], ".")
	zone := awsroute53.HostedZone_FromLookup(broker.Construct, jsii.String("HZone"),
		&awsroute53.HostedZoneProviderProps{
			DomainName: jsii.String(domain),
		},
	)

	broker.dns = awsroute53.NewARecord(broker.Construct, jsii.String("ARecord"),
		&awsroute53.ARecordProps{
			RecordName: jsii.String(host),
			Target: awsroute53.RecordTarget_FromAlias(
				awsroute53targets.NewApiGatewayv2DomainProperties(
					broker.domain.RegionalDomainName(),
					broker.domain.RegionalHostedZoneId(),
				),
			),
			Ttl:  awscdk.Duration_Seconds(jsii.Number(60)),
			Zone: zone,
		},
	)

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
