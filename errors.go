//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import "github.com/fogfish/faults"

const (
	ErrServiceIO = faults.Type("service i/o failed")
	ErrEnqueue   = faults.Type("enqueue is failed")
	ErrDequeue   = faults.Type("dequeue is failed")
)
