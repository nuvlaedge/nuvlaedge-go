package types

import (
	"nuvlaedge-go/nuvlaedge/tests"
	"testing"
)

func TestNewPayloadFromString(t *testing.T) {

	_, err := NewPayloadFromString("test")
	tests.Assert(t, err != nil, "Error should be returned when payload is not a valid JSON")

	p, err := NewPayloadFromString("{\"test\": \"test\"}")
	tests.Assert(t, err == nil, "Error should not be returned when payload is a valid JSON")
	// Payload should be emp
}
