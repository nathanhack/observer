package topic

import (
	"github.com/oklog/ulid"
	"sync"
)

type Topic struct {
	Callbacks    map[ulid.ULID]func(interface{})
	CallbacksMux *sync.RWMutex
	Channels     map[ulid.ULID]chan interface{}
	ChannelsMux  *sync.RWMutex
}

func (t *Topic) RegisterCallback(id ulid.ULID, callback func(data interface{})) {
	t.CallbacksMux.Lock()
	defer t.CallbacksMux.Unlock()

	t.Callbacks[id] = callback
}
func (t *Topic) RegisterChannel(id ulid.ULID, channel chan interface{}) {
	t.ChannelsMux.Lock()
	defer t.ChannelsMux.Unlock()

	t.Channels[id] = channel
}
func (t *Topic) UnregisterCallback(id ulid.ULID) {
	t.CallbacksMux.Lock()
	defer t.CallbacksMux.Unlock()

	if _, has := t.Callbacks[id]; has {
		delete(t.Callbacks, id)
	}
}
func (t *Topic) UnregisterChannel(id ulid.ULID) {
	t.ChannelsMux.Lock()
	defer t.ChannelsMux.Unlock()

	if _, has := t.Channels[id]; has {
		delete(t.Channels, id)
	}
}
func (t *Topic) Send(data interface{}) {
	go func() {
		t.CallbacksMux.RLock()
		defer t.CallbacksMux.RUnlock()

		for _, c := range t.Callbacks {
			go c(data)
		}
	}()

	go func() {
		t.ChannelsMux.RLock()
		defer t.ChannelsMux.RUnlock()

		for _, c := range t.Channels {
			go func() {
				c <- data
			}()
		}
	}()
}
func (t *Topic) SendTo(data interface{}, sendToIDs map[ulid.ULID]bool) {
	go func() {
		t.CallbacksMux.RLock()
		defer t.CallbacksMux.RUnlock()

		for id := range sendToIDs {
			if call, has := t.Callbacks[id]; has {
				go call(data)
			}
		}
	}()

	go func() {
		t.ChannelsMux.RLock()
		defer t.ChannelsMux.RUnlock()

		for id, _ := range sendToIDs {
			if ch, has := t.Channels[id]; has {
				go func() {
					ch <- data
				}()
			}
		}
	}()
}
func (t *Topic) SendExcept(data interface{}, filterOutIDs map[ulid.ULID]bool) {
	go func() {
		t.CallbacksMux.RLock()
		defer t.CallbacksMux.RUnlock()

		for id, call := range t.Callbacks {
			if _, has := filterOutIDs[id]; !has {
				go call(data)
			}
		}
	}()

	go func() {
		t.ChannelsMux.RLock()
		defer t.ChannelsMux.RUnlock()

		for id, ch := range t.Channels {
			if _, has := filterOutIDs[id]; !has {
				go func() {
					ch <- data
				}()
			}
		}
	}()
}
