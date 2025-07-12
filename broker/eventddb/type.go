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
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/listen"
)

const Category = "DynamoDBEventRecord"

// The broker produces only [events.DynamoDBEventRecord], the function is helper.
func Source(q *kernel.ListenerCore) (
	<-chan swarm.Msg[*events.DynamoDBEventRecord],
	chan<- swarm.Msg[*events.DynamoDBEventRecord],
) {
	return listen.Typed[*events.DynamoDBEventRecord](q)
}
