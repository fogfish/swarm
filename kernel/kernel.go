//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

type Kernel struct {
	Emitter  *EmitterIO
	Listener *ListenerIO
}

func New(emitter *EmitterIO, listener *ListenerIO) *Kernel {
	return &Kernel{
		Emitter:  emitter,
		Listener: listener,
	}
}

func (k *Kernel) Close() {
	if k.Listener != nil {
		k.Listener.Close()
	}

	if k.Emitter != nil {
		k.Emitter.Close()
	}
}

func (k *Kernel) Await() {
	if k.Listener != nil {
		k.Listener.Await()
	}

	if k.Emitter != nil {
		k.Emitter.Await()
	}
}
