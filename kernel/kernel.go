//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

type Kernel struct {
	*EmitterKernel
	*ListenerKernel
}

func New(enqueuer *EmitterKernel, dequeuer *ListenerKernel) *Kernel {
	return &Kernel{
		EmitterKernel:  enqueuer,
		ListenerKernel: dequeuer,
	}
}

func (k *Kernel) Close() {
	if k.ListenerKernel != nil {
		k.ListenerKernel.Close()
	}

	if k.EmitterKernel != nil {
		k.EmitterKernel.Close()
	}
}

func (k *Kernel) Await() {
	if k.ListenerKernel != nil {
		k.ListenerKernel.Await()
	}

	if k.EmitterKernel != nil {
		k.EmitterKernel.Await()
	}
}
