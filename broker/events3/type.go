//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events3

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/dequeue"
	"github.com/fogfish/swarm/kernel"
)

const Category = "S3EventRecord"

// The broker produces only [events.S3EventRecord], the function is helper.
func Source(q *kernel.Dequeuer) (
	<-chan swarm.Msg[*events.S3EventRecord],
	chan<- swarm.Msg[*events.S3EventRecord],
) {
	return dequeue.Typed[*events.S3EventRecord](q)
}
