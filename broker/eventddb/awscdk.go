//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventddb

import (
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
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
	Table       awsdynamodb.ITable
	Function    scud.FunctionProps
	EventSource *awslambdaeventsources.DynamoEventSourceProps
}

func (props *SinkProps) assert() {
	if props.Function == nil {
		panic("Function is not defined.")
	}
	if props.Table == nil {
		panic("Table is not defined.")
	}
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	props.assert()

	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}
	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)

	sink.Handler.AddEnvironment(
		jsii.String(EnvConfigSourceDynamoDB),
		props.Table.TableName(),
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

	eventsource := &awslambdaeventsources.DynamoEventSourceProps{
		StartingPosition: awslambda.StartingPosition_LATEST,
	}
	if props.EventSource != nil {
		eventsource = props.EventSource
	}

	source := awslambdaeventsources.NewDynamoEventSource(props.Table, eventsource)

	sink.Handler.AddEventSource(source)

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
// AWS CDK Broker Construct
//
//------------------------------------------------------------------------------

type BrokerProps struct{}

type Broker struct {
	constructs.Construct
	Table awsdynamodb.ITable
}

func NewBroker(scope constructs.Construct, id *string, props *BrokerProps) *Broker {
	broker := &Broker{Construct: constructs.NewConstruct(scope, id)}

	return broker
}

func (broker *Broker) NewTable(props *awsdynamodb.TablePropsV2) awsdynamodb.ITable {
	if props == nil {
		props = &awsdynamodb.TablePropsV2{}
	}

	if props.TableName == nil {
		props.TableName = awscdk.Aws_STACK_NAME()
	}

	if props.PartitionKey == nil && props.SortKey == nil {
		props.PartitionKey = &awsdynamodb.Attribute{
			Type: awsdynamodb.AttributeType_STRING,
			Name: jsii.String("prefix"),
		}

		props.SortKey = &awsdynamodb.Attribute{
			Type: awsdynamodb.AttributeType_STRING,
			Name: jsii.String("suffix"),
		}
	}

	if props.Billing == nil {
		props.Billing = awsdynamodb.Billing_OnDemand(
			&awsdynamodb.MaxThroughputProps{},
		)
	}

	if props.DynamoStream == "" {
		props.DynamoStream = awsdynamodb.StreamViewType_NEW_IMAGE
	}

	broker.Table = awsdynamodb.NewTableV2(broker.Construct, jsii.String("Table"), props)

	return broker.Table
}

func (broker *Broker) AddTable(tableName string) awsdynamodb.ITable {
	broker.Table = awsdynamodb.TableV2_FromTableName(broker.Construct, jsii.String("Table"),
		jsii.String(tableName),
	)

	return broker.Table
}

func (broker *Broker) NewSink(props *SinkProps) *Sink {
	if broker.Table == nil {
		panic("Table is not defined.")
	}

	props.Table = broker.Table

	name := props.Function.UniqueID()
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}

func (broker *Broker) GrantWrite(f awslambda.Function) {
	broker.Table.GrantWriteData(f)

	f.AddEnvironment(
		jsii.String(EnvConfigTargetDynamoDB),
		broker.Table.TableName(),
		nil,
	)
}

func (broker *Broker) GrantRead(f awslambda.Function) {
	broker.Table.GrantReadData(f)

	f.AddEnvironment(
		jsii.String(EnvConfigSourceDynamoDB),
		broker.Table.TableName(),
		nil,
	)
}
