//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

import (
	"os"
	"sync"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/broadcast"
)

var (
	sysf *factory
	once sync.Once
)

func builder() *factory {
	once.Do(func() {
		sysf = &factory{}
		sysf.setup()
	})

	return sysf
}

type factory struct {
	// messaging kernel is running in the externally preemptable mode
	isExternallyPreemptable bool

	ctrlPreempt *broadcast.Broadcaster
}

func (b *factory) setup() {
	if _, has := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); has {
		b.enableExternalPreemption()
	}
}

func (b *factory) enableExternalPreemption() {
	b.isExternallyPreemptable = true
	b.ctrlPreempt = broadcast.New()
}

func (b *factory) Bridge(cfg swarm.Config) *Bridge {
	bridge := newBridge(cfg)
	bridge.ctrlPreempt = b.ctrlPreempt

	return bridge
}

func (b *factory) Emitter(emitter Emitter, cfg swarm.Config) *EmitterCore {
	enqueuer := newEmitter(emitter, cfg)
	enqueuer.ctrlPreempt = b.ctrlPreempt

	return enqueuer
}

func (b *factory) Listener(cathode Listener, config swarm.Config) *ListenerCore {
	dequeuer := newListener(cathode, config)

	return dequeuer
}
