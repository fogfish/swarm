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
