package main

import (
	"testing"
	"time"
)

func TestTimestampInMessageStuct(t *testing.T) {
	s := "test message"

	m := createMessage(s)

	if m.Body != s {
		t.Error("Message supplied not in message body.")
	}

	if m.Timestamp > time.Now().Unix() {
		t.Error("Provided timestamp is in the future.")
	}
}
