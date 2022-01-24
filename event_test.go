package events_test

import (
	"testing"

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
