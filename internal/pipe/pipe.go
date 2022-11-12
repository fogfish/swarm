//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package pipe

import "time"

/*

ForEach applies function for each message in the channel
*/
func ForEach[A any](in <-chan A, f func(A)) {
	go func() {
		var (
			x  A
			ok bool
		)

		for {
			select {
			case x, ok = <-in:
				if !ok {
					return
				}
				f(x)
			}
		}
	}()
}

/*

Emit periodically message from the function
*/
func Emit[T any](eg chan<- T, frequency time.Duration, f func() (T, error)) {
	go func() {
		defer func() {
			// Note: recover from panic on sending to closed channel
			if recover() != nil {
			}
		}()

		for {
			select {
			case <-time.After(frequency):
				if x, err := f(); err == nil {
					eg <- x
				}
			}
		}
	}()
}
