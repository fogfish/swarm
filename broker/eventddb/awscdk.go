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
	Table       awsdynamodb.ITable
	Function    scud.FunctionProps
	EventSource *awslambdaeventsources.DynamoEventSourceProps
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	sink.Handler = scud.NewFunction(sink.Construct, jsii.String("Func"), props.Function)

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

//------------------------------------------------------------------------------
//
// AWS CDK Broker Construct
//
//------------------------------------------------------------------------------

type BrokerProps struct {
	System string
}

type Broker struct {
	constructs.Construct
	Table awsdynamodb.ITable
	acc   int
}

func NewBroker(scope constructs.Construct, id *string, props *BrokerProps) *Broker {
	broker := &Broker{Construct: constructs.NewConstruct(scope, id)}

	return broker
}

// func (broker *Broker) NewTable(props *awsdynamodb.TableProps) awsdynamodb.ITable {
// 	if props == nil {
// 		props = &awsdynamodb.TableProps{}
// 	}

// 	if props.TableName == nil {
// 		props.TableName = awscdk.Aws_STACK_NAME()
// 	}

// 	if props.PartitionKey == nil && props.SortKey == nil {
// 		props.PartitionKey = &awsdynamodb.Attribute{
// 			Type: awsdynamodb.AttributeType_STRING,
// 			Name: jsii.String("prefix"),
// 		}

// 		props.SortKey = &awsdynamodb.Attribute{
// 			Type: awsdynamodb.AttributeType_STRING,
// 			Name: jsii.String("suffix"),
// 		}
// 	}

// 	if props.BillingMode == "" {
// 		props.BillingMode = awsdynamodb.BillingMode_PAY_PER_REQUEST
// 	}

// 	if props.Stream == "" {
// 		props.Stream = awsdynamodb.StreamViewType_NEW_IMAGE
// 	}

// 	broker.Table = awsdynamodb.NewTable(broker.Construct, jsii.String("Table"), props)

// 	return broker.Table
// }

// func (broker *Broker) AddTable(tableName string) awsdynamodb.ITable {
// 	broker.Table = awsdynamodb.Table_FromTableName(broker.Construct, jsii.String("Table"),
// 		jsii.String(tableName),
// 	)

// 	return broker.Table
// }

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
		props.Billing = awsdynamodb.Billing_OnDemand()
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

	broker.acc++
	name := "Sink" + strconv.Itoa(broker.acc)
	sink := NewSink(broker.Construct, jsii.String(name), props)

	return sink
}
