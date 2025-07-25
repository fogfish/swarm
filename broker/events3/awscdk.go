//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events3

import (
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
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

type Sink struct {
	constructs.Construct
	Handler awslambda.Function
}

type SinkProps struct {
	Agent       *string
	Bucket      awss3.Bucket
	EventSource *awslambdaeventsources.S3EventSourceProps
	Function    scud.FunctionProps
}

func (props *SinkProps) assert() {
	if props.Function == nil {
		panic("Function is not defined.")
	}
	if props.Bucket == nil {
		panic("Bucket is not defined.")
	}
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	props.assert()

	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}
	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)

	sink.Handler.AddEnvironment(
		jsii.String(EnvConfigSourceS3),
		props.Bucket.BucketName(),
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

	eventsource := &awslambdaeventsources.S3EventSourceProps{
		Events: &[]awss3.EventType{
			awss3.EventType_OBJECT_CREATED,
			awss3.EventType_OBJECT_REMOVED,
		},
	}
	if props.EventSource != nil {
		eventsource = props.EventSource
	}

	sink.Handler.AddEventSource(
		awslambdaeventsources.NewS3EventSource(props.Bucket, eventsource),
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
	Bucket awss3.Bucket
}

func NewBroker(scope constructs.Construct, id *string, props *BrokerProps) *Broker {
	broker := &Broker{Construct: constructs.NewConstruct(scope, id)}

	return broker
}

func (broker *Broker) NewBucket(props *awss3.BucketProps) awss3.Bucket {
	if props == nil {
		props = &awss3.BucketProps{}
	}

	if props.BucketName == nil {
		props.BucketName = awscdk.Aws_STACK_NAME()
	}

	broker.Bucket = awss3.NewBucket(broker.Construct, jsii.String("Bucket"), props)

	return broker.Bucket
}

func (broker *Broker) NewSink(props *SinkProps) *Sink {
	if broker.Bucket == nil {
		panic("Bucket is not defined.")
	}

	props.Bucket = broker.Bucket

	name := props.Function.UniqueID()
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}

func (broker *Broker) GrantWrite(f awslambda.Function) {
	broker.Bucket.GrantWrite(f, nil, nil)

	f.AddEnvironment(
		jsii.String(EnvConfigTargetS3),
		broker.Bucket.BucketName(),
		nil,
	)
}

func (broker *Broker) GrantRead(f awslambda.Function) {
	broker.Bucket.GrantRead(f, nil)

	f.AddEnvironment(
		jsii.String(EnvConfigSourceS3),
		broker.Bucket.BucketName(),
		nil,
	)
}
