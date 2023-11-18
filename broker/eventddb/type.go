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
	"github.com/fogfish/swarm"
	queue "github.com/fogfish/swarm/queue/events"
)

const Category = "eventddb.Event"

type Event swarm.Event[*events.DynamoDBEventRecord]

func (Event) HKT1(swarm.EventType)             {}
func (Event) HKT2(*events.DynamoDBEventRecord) {}

func Dequeue(q swarm.Broker) (<-chan *Event, chan<- *Event) {
	return queue.Dequeue[*events.DynamoDBEventRecord, Event](q)
}
