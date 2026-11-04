package llm

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestDeepSeekGenerateResponse(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            t.Errorf("method = %s, want POST", r.Method)
        }
        if r.URL.Path != "/chat/completions" {
            t.Errorf("path = %s, want /chat/completions", r.URL.Path)
        }
        if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
            t.Errorf("authorization = %q, want %q", got, "Bearer test-key")
        }

        var request deepSeekRequest
        if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            t.Errorf("failed to decode request: %v", err)
        }
        if request.Model != DefaultDeepSeekModel {
            t.Errorf("model = %q, want %q", request.Model, DefaultDeepSeekModel)
        }
        if len(request.Messages) != 2 {
            t.Fatalf("messages = %d, want 2", len(request.Messages))
        }
        if request.Messages[0].Role != "system" || request.Messages[0].Content != "system prompt" {
            t.Errorf("system message = %#v", request.Messages[0])
        }
        if request.Messages[1].Role != "user" || request.Messages[1].Content != "user prompt" {
            t.Errorf("user message = %#v", request.Messages[1])
        }

        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(map[string]any{
            "choices": []map[string]any{
                {
                    "message": map[string]string{
                        "role":    "assistant",
                        "content": " hello ",
                    },
                },
            },
        })
    }))
    defer server.Close()

    client := NewDeepSeekClient("test-key", "", server.URL)
    response, err := client.GenerateResponse(context.Background(), "system prompt", "user prompt")
    if err != nil {
        t.Fatalf("GenerateResponse failed: %v", err)
    }
    if response != "hello" {
        t.Errorf("response = %q, want %q", response, "hello")
    }
}
