package observer

import (
	"github.com/oklog/ulid"
	"reflect"
	"testing"
	"time"
)

func TestObserver_Send(t *testing.T) {
	topicName := "name"
	expected := "test"
	callbackCalled := false
	callback := func(data interface{}) {
		if !reflect.DeepEqual(data, expected) {
			t.Errorf("expected %v but found %v", expected, data)
		}
		callbackCalled = true
	}

	channel := make(chan interface{}, 10)

	o := New()
	id := CreateID()

	o.RegisterCallback(id, topicName, callback)
	o.RegisterChannel(id, topicName, channel)

	o.Send(topicName, expected)

	data := <-channel
	if !reflect.DeepEqual(expected, data) {
		t.Errorf("expected %v but found %v", expected, data)
	}

	time.Sleep(500 * time.Millisecond)
	if !callbackCalled {
		t.Errorf("expected callback to be called")
	}
	callbackCalled = false
	o.SendExcept(topicName, expected, map[ulid.ULID]bool{id: true})

	select {
	case <-channel:
		t.Errorf("expected no channel data")
	case <-time.After(1 * time.Second):
	}

	if callbackCalled {
		t.Errorf("expected no callback to be called")
	}
}
