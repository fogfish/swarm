//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package pipe

import (
	"os"
	"sync"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/emit"
	"github.com/fogfish/swarm/examples/traceroute/core"
	"github.com/fogfish/swarm/kernel"
	"github.com/fogfish/swarm/listen"
)

var (
	sqsEmit *kernel.EmitterIO
	sqsOnce sync.Once

	sqsRecv *kernel.ListenerIO
	sqsEcno sync.Once
)

func getSqsEmitter(q string) *kernel.EmitterIO {
	sqsOnce.Do(func() {
		sqsEmit = sqs.Must(
			sqs.Emitter().Build(q),
		)
	})
	return sqsEmit
}

func getSqsListener(q string) *kernel.ListenerIO {
	sqsEcno.Do(func() {
		sqsRecv = sqs.Must(
			sqs.Listener().Build(q),
		)
		go sqsRecv.Await()
	})
	return sqsRecv
}

func ToSQS() chan<- core.ReqTrace {
	return swarm.LogDeadLetters(
		emit.Event[core.ReqTrace](
			getSqsEmitter(os.Getenv(sqs.EnvConfigTargetSQS)),
		),
	)
}

func FromSQS() (<-chan core.ReqTrace, chan<- core.ReqTrace) {
	return listen.Event[core.ReqTrace](
		getSqsListener(os.Getenv(sqs.EnvConfigSourceSQS)),
	)
}
