//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package broadcast

import (
	"context"
	"sync"
)

type Broadcaster struct {
	mu sync.RWMutex
	ch []chan chan struct{}
}

func New() *Broadcaster {
	return &Broadcaster{
		ch: make([]chan chan struct{}, 0),
	}
}

func (b *Broadcaster) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, ch := range b.ch {
		close(ch)
	}

	b.ch = nil
}

func (b *Broadcaster) Register() chan chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan chan struct{}, 1)
	b.ch = append(b.ch, ch)

	return ch
}

func (b *Broadcaster) Unregister(ch chan chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, c := range b.ch {
		if c == ch {
			b.ch = append(b.ch[:i], b.ch[i+1:]...)
			close(ch)
			return
		}
	}
}

func (b *Broadcaster) Cast(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.ch) == 0 {
		return nil
	}

	ackCh := make(chan struct{}, len(b.ch))
	defer close(ackCh)

	for _, ch := range b.ch {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- ackCh:
		default:
		}
	}

	for range len(b.ch) {
		select {
		case <-ackCh:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
