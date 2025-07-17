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
	"github.com/fogfish/swarm/broker/websocket"
	"github.com/fogfish/swarm/examples/traceroute/core"
	"github.com/fogfish/swarm/kernel"
)

var (
	wsEmit *kernel.EmitterIO
	wsOnce sync.Once
)

func getWebSocketEmitter() *kernel.EmitterIO {
	wsOnce.Do(func() {
		wsEmit = websocket.Must(
			websocket.Emitter().Build(
				os.Getenv(websocket.EnvConfigTargetWebSocketUrl),
			),
		)
	})
	return wsEmit
}

func ActStatus() chan<- core.ActStatus {
	return swarm.LogDeadLetters(
		websocket.EmitEvent[core.ActStatus](
			getWebSocketEmitter(),
		),
	)
}
