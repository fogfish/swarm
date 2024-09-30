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
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/fogfish/scud"
)

//------------------------------------------------------------------------------
//
// AWS CDK Sink Construct
//
//------------------------------------------------------------------------------

type Sink struct {
	constructs.Construct
	Handler awslambda.IFunction
}

type SinkProps struct {
	Bucket      awss3.Bucket
	EventSource *awslambdaeventsources.S3EventSourceProps
	Function    scud.FunctionProps
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)

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
	Bucket awss3.Bucket
	acc    int
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

	broker.acc++
	name := "Sink" + strconv.Itoa(broker.acc)
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}
