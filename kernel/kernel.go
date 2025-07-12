//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

type Kernel struct {
	*EmitterCore
	*ListenerCore
}

func New(enqueuer *EmitterCore, dequeuer *ListenerCore) *Kernel {
	return &Kernel{
		EmitterCore:  enqueuer,
		ListenerCore: dequeuer,
	}
}

func (k *Kernel) Close() {
	if k.ListenerCore != nil {
		k.ListenerCore.Close()
	}

	if k.EmitterCore != nil {
		k.EmitterCore.Close()
	}
}

func (k *Kernel) Await() {
	if k.ListenerCore != nil {
		k.ListenerCore.Await()
	}

	if k.EmitterCore != nil {
		k.EmitterCore.Await()
	}
}
