package events_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kmalaver/events"
)

type St struct {
	Name string
}

func TestNewEvent(t *testing.T) {
	eString := events.New[string]()
	if eString == nil {
		t.Error("Expected event to be created")
	}
}

func TestAsyncEvents(t *testing.T) {
	event := events.New[int]()
	event.SubscribeAsync(func(data int) {
		fmt.Println("A: ", data)
	})
	event.SubscribeAsync(func(data int) {
		time.Sleep(time.Second)
		fmt.Println("B: ", data)
	})
	event.SubscribeAsync(func(data int) {
		fmt.Println("C: ", data)
	})
	event.SubscribeAsync(func(data int) {
		fmt.Println("D: ", data)
	})

	go event.Dispatch(1)
	event.Dispatch(2)
	event.Wait()
}
