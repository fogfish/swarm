//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package system

import "sync"

type Channels struct {
	sync.Mutex
	channels map[string]Closer
}

func NewChannels() *Channels {
	return &Channels{
		channels: make(map[string]Closer),
	}
}

func (chs *Channels) Length() int {
	return len(chs.channels)
}

func (chs *Channels) Attach(id string, ch Closer) {
	chs.Lock()
	defer chs.Unlock()

	chs.channels[id] = ch
}

func (chs *Channels) Sync() {
	for _, ch := range chs.channels {
		ch.Sync()
	}
}

func (chs *Channels) Close() {
	chs.Lock()
	defer chs.Unlock()

	for _, ch := range chs.channels {
		ch.Close()
	}

	chs.channels = make(map[string]Closer)
}
