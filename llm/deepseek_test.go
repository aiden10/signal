package llm

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "signal/models"
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
    response, err := client.GenerateResponse(context.Background(), "system prompt", "user prompt", nil)
    if err != nil {
        t.Fatalf("GenerateResponse failed: %v", err)
    }
    if response != "hello" {
        t.Errorf("response = %q, want %q", response, "hello")
    }
}

func TestDeepSeekGenerateResponseWithImage(t *testing.T) {
    var received deepSeekRequest
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
            t.Errorf("failed to decode request: %v", err)
        }

        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(map[string]any{
            "choices": []map[string]any{
                {
                    "message": map[string]string{
                        "role":    "assistant",
                        "content": "a picture",
                    },
                },
            },
        })
    }))
    defer server.Close()

    client := NewDeepSeekClient("test-key", "", server.URL)
    _, err := client.GenerateResponse(context.Background(), "", "what is this", []models.Image{
        {ContentType: "image/png", Data: []byte("abc")},
    })
    if err != nil {
        t.Fatalf("GenerateResponse failed: %v", err)
    }

    if len(received.Messages) != 1 {
        t.Fatalf("messages = %d, want 1", len(received.Messages))
    }

    parts, ok := received.Messages[0].Content.([]any)
    if !ok {
        t.Fatalf("message content = %#v, want content parts", received.Messages[0].Content)
    }
    if len(parts) != 2 {
        t.Fatalf("parts = %d, want 2", len(parts))
    }

    textPart, ok := parts[0].(map[string]any)
    if !ok || textPart["type"] != "text" || textPart["text"] != "what is this" {
        t.Errorf("text part = %#v", parts[0])
    }

    imagePart, ok := parts[1].(map[string]any)
    if !ok || imagePart["type"] != "image_url" {
        t.Fatalf("image part = %#v", parts[1])
    }
    imageURL, ok := imagePart["image_url"].(map[string]any)
    if !ok || imageURL["url"] != "data:image/png;base64,YWJj" {
        t.Errorf("image url = %#v", imagePart["image_url"])
    }
}
