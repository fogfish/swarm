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
	"github.com/fogfish/swarm/broker/eventbridge"
	"github.com/fogfish/swarm/emit"
	"github.com/fogfish/swarm/examples/traceroute/core"
	"github.com/fogfish/swarm/kernel"
)

var (
	ebEmit *kernel.EmitterIO
	ebOnce sync.Once
)

func getEventBridgeEmitter() *kernel.EmitterIO {
	ebOnce.Do(func() {
		ebEmit = eventbridge.Must(
			eventbridge.Emitter().Build(
				os.Getenv(eventbridge.EnvConfigTargetEventBus),
			),
		)
	})
	return ebEmit
}

func ToEventBridge() chan<- core.ReqTrace {
	return swarm.LogDeadLetters(
		emit.Event[core.ReqTrace](
			getEventBridgeEmitter(),
		),
	)
}
