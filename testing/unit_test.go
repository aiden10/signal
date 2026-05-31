package testing

import (
	"testing"
)

func TestGenerateEmbedding(t *testing.T) {
	s := newTestSuite(t)
	vec, err := s.LLM.GenerateEmbedding(Data.TestMessageGeneric)
	if err != nil {
        t.Fatalf("TestGenerateEmbedding failed: %v", err)
    }
	t.Logf("Generated vector: %v", vec)
}