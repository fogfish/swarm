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

	ctrlPreempt chan chan struct{}
}

func (b *factory) setup() {
	if _, has := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); has {
		b.enableExternalPreemption()
	}
}

func (b *factory) enableExternalPreemption() {
	b.isExternallyPreemptable = true
	b.ctrlPreempt = make(chan chan struct{}, 1)
}

func (b *factory) Bridge(cfg swarm.Config) *Bridge {
	bridge := newBridge(cfg)
	bridge.ctrlPreempt = b.ctrlPreempt

	return bridge
}

func (b *factory) Enqueuer(emitter Emitter, cfg swarm.Config) *Enqueuer {
	enqueuer := newEnqueuer(emitter, cfg)
	enqueuer.ctrlPreempt = b.ctrlPreempt

	return enqueuer
}

func (b *factory) Dequeuer(cathode Cathode, config swarm.Config) *Dequeuer {
	dequeuer := newDequeuer(cathode, config)

	return dequeuer
}
