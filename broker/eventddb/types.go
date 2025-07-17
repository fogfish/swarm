//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventddb

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fogfish/faults"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/listen"
)

const Category = "DynamoDBEventRecord"

// The broker produces only [events.DynamoDBEventRecord], the function is helper.
func Listen(q *kernel.ListenerIO) (
	<-chan swarm.Msg[*events.DynamoDBEventRecord],
	chan<- swarm.Msg[*events.DynamoDBEventRecord],
) {
	return listen.Typed[*events.DynamoDBEventRecord](q)
}

func UnmarshalMap(obj map[string]events.DynamoDBAttributeValue, out any) error {
	avobj, err := ToAttributeValueMap(obj)
	if err != nil {
		return err
	}

	return attributevalue.UnmarshalMap(avobj, out)
}

const (
	errCodecKey     = faults.Safe1[string]("failed to convert attribute <%s>")
	errCodecSeq     = faults.Safe1[int]("failed to convert attribute at %d")
	errCodecUnknown = faults.Safe1[events.DynamoDBDataType]("unknown DynamoDB attribute type: %v")
)

// AWS Lambda Events uses events.DynamoDBAttributeValue to encode DynamoDB objects.
// It is not compatible with the AWS SDK v2 types.AttributeValue so that no efficient encoders exists.
// This function converts events.DynamoDBAttributeValue to types.AttributeValue allowing further processing with the AWS SDK v2.
func ToAttributeValueMap(obj map[string]events.DynamoDBAttributeValue) (map[string]types.AttributeValue, error) {
	avobj := make(map[string]types.AttributeValue, len(obj))

	for key, av := range obj {
		value, err := toAttributeValue(av)
		if err != nil {
			return nil, errCodecKey.With(err, key)
		}
		avobj[key] = value
	}

	return avobj, nil
}

// toAttributeValue converts Lambda events.DynamoDBAttributeValue to SDK v2 types.AttributeValue
func toAttributeValue(av events.DynamoDBAttributeValue) (types.AttributeValue, error) {
	switch av.DataType() {
	case events.DataTypeString:
		return &types.AttributeValueMemberS{Value: av.String()}, nil

	case events.DataTypeNumber:
		return &types.AttributeValueMemberN{Value: av.Number()}, nil

	case events.DataTypeBoolean:
		return &types.AttributeValueMemberBOOL{Value: av.Boolean()}, nil

	case events.DataTypeBinary:
		return &types.AttributeValueMemberB{Value: av.Binary()}, nil

	case events.DataTypeNull:
		return &types.AttributeValueMemberNULL{Value: true}, nil

	case events.DataTypeStringSet:
		return &types.AttributeValueMemberSS{Value: av.StringSet()}, nil

	case events.DataTypeNumberSet:
		return &types.AttributeValueMemberNS{Value: av.NumberSet()}, nil

	case events.DataTypeBinarySet:
		return &types.AttributeValueMemberBS{Value: av.BinarySet()}, nil

	case events.DataTypeList:
		seq := av.List()
		avseq := make([]types.AttributeValue, len(seq))
		for i, item := range seq {
			val, err := toAttributeValue(item)
			if err != nil {
				return nil, errCodecSeq.With(err, i)
			}
			avseq[i] = val
		}
		return &types.AttributeValueMemberL{Value: avseq}, nil

	case events.DataTypeMap:
		obj := av.Map()
		avobj := make(map[string]types.AttributeValue, len(obj))
		for key, value := range obj {
			val, err := toAttributeValue(value)
			if err != nil {
				return nil, errCodecKey.With(err, key)
			}
			avobj[key] = val
		}
		return &types.AttributeValueMemberM{Value: avobj}, nil

	default:
		return nil, errCodecUnknown.With(nil, av.DataType())
	}
}
