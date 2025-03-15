//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"fmt"
	"time"

	"github.com/fogfish/faults"
)

const (
	ErrServiceIO  = faults.Type("service i/o failed")
	ErrEnqueue    = faults.Type("enqueue i/o failed")
	ErrEncoder    = faults.Type("encoder failure")
	ErrDequeue    = faults.Type("dequeue i/o failed")
	ErrDecoder    = faults.Type("decoder failure")
	ErrRouting    = faults.Safe1[string]("routing has failed (cat %s)")
	ErrCatUnknown = faults.Safe1[string]("unknown category %s")
)

type errTimeout struct {
	op    string
	timer time.Duration
}

func ErrTimeout(op string, timer time.Duration) error {
	return errTimeout{
		op:    op,
		timer: timer,
	}
}

func (err errTimeout) Error() string {
	return fmt.Sprintf("timeout %s after %s", err.op, err.timer)
}

func (err errTimeout) Timeout() time.Duration {
	return err.timer
}
