package observer

import (
	"github.com/nathanhack/observer/internal/topic"
	"github.com/oklog/ulid"
	"math/rand"
	"sync"
	"time"
)

type Observer interface {
	RegisterCallback(id ulid.ULID, topicName string, callback func(data interface{}))
	RegisterChannel(id ulid.ULID, topicName string, channel chan interface{})

	UnregisterCallback(id ulid.ULID, topicName string)
	UnregisterChannel(id ulid.ULID, topicName string)

	Send(topicName string, data interface{})
	SendTo(topicName string, data interface{}, sendToIDs map[ulid.ULID]bool)
	SendExcept(topicName string, data interface{}, filterOutIDs map[ulid.ULID]bool)
}

func CreateID() ulid.ULID {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy)
}

type observer struct {
	topics map[string]*topic.Topic
	mux    *sync.RWMutex
}

func New() Observer {
	return &observer{
		topics: map[string]*topic.Topic{},
		mux:    &sync.RWMutex{},
	}
}

func (o *observer) RegisterCallback(id ulid.ULID, topicName string, callback func(data interface{})) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if _, has := o.topics[topicName]; !has {
		o.topics[topicName] = &topic.Topic{
			Callbacks:    map[ulid.ULID]func(interface{}){},
			CallbacksMux: &sync.RWMutex{},
			Channels:     map[ulid.ULID]chan interface{}{},
			ChannelsMux:  &sync.RWMutex{},
		}
	}
	o.topics[topicName].RegisterCallback(id, callback)
}

func (o *observer) RegisterChannel(id ulid.ULID, topicName string, channel chan interface{}) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if _, has := o.topics[topicName]; !has {
		o.topics[topicName] = &topic.Topic{
			Callbacks:    map[ulid.ULID]func(interface{}){},
			CallbacksMux: &sync.RWMutex{},
			Channels:     map[ulid.ULID]chan interface{}{},
			ChannelsMux:  &sync.RWMutex{},
		}
	}

	o.topics[topicName].RegisterChannel(id, channel)
}

func (o *observer) UnregisterCallback(id ulid.ULID, topicName string) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if x, has := o.topics[topicName]; has {
		x.UnregisterCallback(id)
	}
}

func (o *observer) UnregisterChannel(id ulid.ULID, topicName string) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if x, has := o.topics[topicName]; has {
		x.UnregisterChannel(id)
	}
}

func (o *observer) Send(topicName string, data interface{}) {
	o.mux.RLock()
	defer o.mux.RUnlock()

	if x, has := o.topics[topicName]; has {
		go x.Send(data)
	}
}

func (o *observer) SendTo(topicName string, data interface{}, sendToIDs map[ulid.ULID]bool) {
	o.mux.RLock()
	defer o.mux.RUnlock()

	if x, has := o.topics[topicName]; has {
		go x.SendTo(data, sendToIDs)
	}
}

func (o *observer) SendExcept(topicName string, data interface{}, filterOutIDs map[ulid.ULID]bool) {
	o.mux.RLock()
	defer o.mux.RUnlock()

	if x, has := o.topics[topicName]; has {
		go x.SendExcept(data, filterOutIDs)
	}
}
