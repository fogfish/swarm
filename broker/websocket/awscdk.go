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
	"github.com/fogfish/swarm"
)

//------------------------------------------------------------------------------
//
// AWS CDK Sink Construct
//
//------------------------------------------------------------------------------

const stage = "ws"

type Sink struct {
	constructs.Construct
	Handler awslambda.Function
}

type SinkProps struct {
	Agent *string

	// Category of event to process. It is used to create a WebSocket route.
	Category string

	// Properties of Lambda function to handle the category
	Function scud.FunctionProps

	// WebSocket Api Gateway lambda is bound with
	Gateway awsapigatewayv2.WebSocketApi
}

func (props *SinkProps) assert() {
	if props.Function == nil {
		panic("Function is not defined.")
	}
	if props.Gateway == nil {
		panic("Gateway is not defined.")
	}
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	props.assert()

	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)
	sink.Handler.AddEnvironment(
		jsii.String(EnvConfigEventCategory),
		jsii.String(props.Category),
		nil,
	)

	sink.Handler.AddEnvironment(
		jsii.String(EnvConfigSourceWebSocketUrl),
		jsii.String(aws.ToString(props.Gateway.ApiEndpoint())+"/"+stage),
		nil,
	)

	if props.Agent != nil {
		sink.Handler.AddEnvironment(
			jsii.String(swarm.EnvConfigAgent),
			props.Agent,
			nil,
		)
	}

	if ttf := timeToFlight(props.Function); ttf != nil {
		tsf := ttf.ToSeconds(nil)
		tsi := int(aws.ToFloat64(tsf))

		sink.Handler.AddEnvironment(
			jsii.String(swarm.EnvConfigTimeToFlight),
			jsii.String(strconv.Itoa(tsi)),
			nil,
		)
	}

	it := integrations.NewWebSocketLambdaIntegration(jsii.String(props.Category), sink.Handler,
		&integrations.WebSocketLambdaIntegrationProps{},
	)

	props.Gateway.AddRoute(jsii.String(props.Category),
		&awsapigatewayv2.WebSocketRouteOptions{
			Integration: it,
		},
	)

	return sink
}

func timeToFlight(props scud.FunctionProps) awscdk.Duration {
	if props == nil {
		return nil
	}

	switch v := props.(type) {
	case *scud.FunctionGoProps:
		if v.FunctionProps != nil && v.Timeout != nil {
			return v.Timeout
		}
	case *scud.ContainerGoProps:
		if v.DockerImageFunctionProps != nil && v.Timeout != nil {
			return v.Timeout
		}
	}

	return nil
}

//------------------------------------------------------------------------------
//
// AWS CDK Stack Construct
//
//------------------------------------------------------------------------------

type BrokerProps struct{}

type Broker struct {
	constructs.Construct
	Gateway    awsapigatewayv2.WebSocketApi
	Authorizer awsapigatewayv2.IWebSocketRouteAuthorizer
	domain     awsapigatewayv2.DomainName
	dns        awsroute53.ARecord
}

func NewBroker(scope constructs.Construct, id *string, props *BrokerProps) *Broker {
	broker := &Broker{Construct: constructs.NewConstruct(scope, id)}

	return broker
}

// Grant permission to write events to the WebSocket
func (broker *Broker) GrantWrite(f awslambda.Function) {
	broker.Gateway.GrantManageConnections(f)

	f.AddEnvironment(
		jsii.String(EnvConfigTargetWebSocketUrl),
		jsii.String(aws.ToString(broker.Gateway.ApiEndpoint())+"/"+stage),
		nil,
	)
}

// Grant permission to read events from the EventBus.
func (broker *Broker) GrantRead(f awslambda.Function) {
	f.AddEnvironment(
		jsii.String(EnvConfigSourceWebSocketUrl),
		jsii.String(aws.ToString(broker.Gateway.ApiEndpoint())+"/"+stage),
		nil,
	)
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

	name := props.Function.UniqueID()
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}
