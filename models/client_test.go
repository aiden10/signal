package models

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestSignalClientSendMessageReturnsTimestamp(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/v2/send" {
            t.Errorf("path = %q, want %q", r.URL.Path, "/v2/send")
        }

        var body struct {
            Message    string   `json:"message"`
            Number     string   `json:"number"`
            Recipients []string `json:"recipients"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            t.Errorf("failed to decode body: %v", err)
        }
        if body.Message != "hello" || body.Number != "+15550001111" || len(body.Recipients) != 1 {
            t.Errorf("body = %#v", body)
        }

        w.WriteHeader(http.StatusCreated)
        _ = json.NewEncoder(w).Encode(map[string]string{"timestamp": "12345"})
    }))
    defer server.Close()

    client := NewSignalClient(server.URL, "+15550001111", "", log.New(io.Discard, "", 0))
    timestamp, err := client.SendMessage("hello", "group")
    if err != nil {
        t.Fatalf("SendMessage failed: %v", err)
    }
    if timestamp != "12345" {
        t.Errorf("timestamp = %q, want %q", timestamp, "12345")
    }
}

func TestSignalClientGetAttachment(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/v1/attachments/image.jpg" {
            t.Errorf("path = %q, want %q", r.URL.Path, "/v1/attachments/image.jpg")
        }
        _, _ = w.Write([]byte("image bytes"))
    }))
    defer server.Close()

    client := NewSignalClient(server.URL, "+15550001111", "", log.New(io.Discard, "", 0))
    data, err := client.GetAttachment("image.jpg")
    if err != nil {
        t.Fatalf("GetAttachment failed: %v", err)
    }
    if string(data) != "image bytes" {
        t.Errorf("data = %q, want %q", string(data), "image bytes")
    }
}
