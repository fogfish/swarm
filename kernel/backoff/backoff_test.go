//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package backoff_test

import (
	"errors"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm/kernel/backoff"
)

func TestConst(t *testing.T) {
	seq := backoff.Const(1*time.Millisecond, 3).Seq()

	it.Then(t).Should(
		it.Equal(len(seq), 3),
		it.Equal(seq[0], seq[1]),
		it.Equal(seq[1], seq[2]),
	)
}

func TestLinera(t *testing.T) {
	seq := backoff.Linear(1*time.Millisecond, 3).Seq()

	it.Then(t).Should(
		it.Equal(len(seq), 3),
		it.Less(seq[0], seq[1]),
		it.Less(seq[1], seq[2]),
	)
}

func TestExp(t *testing.T) {
	seq := backoff.Exp(1*time.Millisecond, 3, 0.5).Seq()

	it.Then(t).Should(
		it.Equal(len(seq), 3),
		it.Less(seq[0], seq[1]),
		it.Less(seq[1], seq[2]),
	)
}

func TestDeadline(t *testing.T) {
	seq := backoff.
		Const(1*time.Millisecond, 10).
		Deadline(5 * time.Millisecond).
		Seq()

	it.Then(t).Should(
		it.Equal(len(seq), 5),
	)
}

func TestRetrySuccess(t *testing.T) {
	n := 0

	err := backoff.Const(1*time.Millisecond, 3).Retry(
		func() error {
			n = n + 1
			return nil
		},
	)

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(n, 1),
	)
}

func TestRetryFail(t *testing.T) {
	n := 0

	err := backoff.Const(1*time.Millisecond, 3).Retry(
		func() error {
			n = n + 1
			return errors.New("skip")
		},
	)

	it.Then(t).ShouldNot(
		it.Nil(err),
	).Should(
		it.Equal(n, 3),
	)
}
