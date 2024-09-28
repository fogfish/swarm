//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package kernel

type Kernel struct {
	*Enqueuer
	*Dequeuer
}

func New(enqueuer *Enqueuer, dequeuer *Dequeuer) *Kernel {
	return &Kernel{
		Enqueuer: enqueuer,
		Dequeuer: dequeuer,
	}
}

func (k *Kernel) Close() {
	if k.Dequeuer != nil {
		k.Dequeuer.Close()
	}

	if k.Enqueuer != nil {
		k.Enqueuer.Close()
	}
}

func (k *Kernel) Await() {
	if k.Dequeuer != nil {
		k.Dequeuer.Await()
	}

	if k.Enqueuer != nil {
		k.Enqueuer.Await()
	}
}
