package testing

import (
	"testing"
)

func TestHandleDataMessage(t *testing.T) {
	s := newTestSuite(t)
	err := s.EventHandler.HandleDataMessage(s.Phone, Data.Username, Data.GeminiPromptMessage, true)
	if err != nil {
        t.Fatalf("HandleDataMessage failed: %v", err)
    }
}