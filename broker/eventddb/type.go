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
	queue "github.com/fogfish/swarm/queue"
)

const Category = "DynamoDBEventRecord"

func Dequeue(q swarm.Broker) (<-chan swarm.Msg[*events.DynamoDBEventRecord], chan<- swarm.Msg[*events.DynamoDBEventRecord]) {
	return queue.Dequeue[*events.DynamoDBEventRecord](q)
}
