package handlers

import (
    "path/filepath"
    "strings"
    "testing"

    "signal/db"
    "signal/models"
)

type recordingLLM struct {
    responseCalls int
    images        []models.Image
    prompts       []string
}

func (l *recordingLLM) GenerateResponse(relevantMessages, recentMessages []models.Message, images []models.Image, systemPrompt, initialPrompt string) string {
    l.responseCalls++
    l.images = images
    l.prompts = append(l.prompts, initialPrompt)
    return "bot response"
}

func (l *recordingLLM) GenerateEmbedding(text string) ([]float32, error) {
    return []float32{1, 0}, nil
}

func newEventTestHandler(t *testing.T) (*EventHandler, *recordingSender, *recordingLLM) {
    t.Helper()

    database, err := db.NewDatabase(filepath.Join(t.TempDir(), "test.db"))
    if err != nil {
        t.Fatalf("failed to open test database: %v", err)
    }
    t.Cleanup(func() {
        _ = database.Close()
    })

    if err := database.Migrate(); err != nil {
        t.Fatalf("failed to migrate test database: %v", err)
    }

    history := NewHistoryHandler(database)
    settings := NewSettingsHandler(database)
    sender := &recordingSender{
        timestamp:   1000,
        attachments: map[string][]byte{},
    }
    llmClient := &recordingLLM{}
    return NewEventHandler(history, settings, llmClient, sender, "", ""), sender, llmClient
}

func TestReplyToBotTriggersAI(t *testing.T) {
    handler, sender, llmClient := newEventTestHandler(t)

    if err := handler.HandleIncomingMessage(models.IncomingMessage{
        GroupID:   "group",
        Author:    "Alice",
        Text:      "@gemini hello",
        Timestamp: 1,
    }, true); err != nil {
        t.Fatalf("HandleIncomingMessage failed: %v", err)
    }
    if len(sender.messages) != 1 {
        t.Fatalf("sent %d messages, want 1", len(sender.messages))
    }

    if err := handler.HandleIncomingMessage(models.IncomingMessage{
        GroupID:   "group",
        Author:    "Alice",
        Text:      "tell me more",
        Timestamp: 2,
        Quote: &models.Quote{
            ID:   1001,
            Text: "bot response",
        },
    }, true); err != nil {
        t.Fatalf("HandleIncomingMessage failed: %v", err)
    }

    if llmClient.responseCalls != 2 {
        t.Errorf("response calls = %d, want 2", llmClient.responseCalls)
    }
    if len(sender.messages) != 2 {
        t.Fatalf("sent %d messages, want 2", len(sender.messages))
    }
    if !strings.Contains(llmClient.prompts[1], "Your earlier message: bot response") {
        t.Errorf("prompt = %q, want quoted bot message", llmClient.prompts[1])
    }
}

func TestReplyToUserMessageDoesNotTriggerAI(t *testing.T) {
    handler, sender, llmClient := newEventTestHandler(t)

    if err := handler.HandleIncomingMessage(models.IncomingMessage{
        GroupID:   "group",
        Author:    "Alice",
        Text:      "hello",
        Timestamp: 1,
    }, true); err != nil {
        t.Fatalf("HandleIncomingMessage failed: %v", err)
    }

    if err := handler.HandleIncomingMessage(models.IncomingMessage{
        GroupID:   "group",
        Author:    "Bob",
        Text:      "replying to Alice",
        Timestamp: 2,
        Quote: &models.Quote{
            ID:   1,
            Text: "hello",
        },
    }, true); err != nil {
        t.Fatalf("HandleIncomingMessage failed: %v", err)
    }

    if llmClient.responseCalls != 0 {
        t.Errorf("response calls = %d, want 0", llmClient.responseCalls)
    }
    if len(sender.messages) != 0 {
        t.Errorf("sent %d messages, want 0", len(sender.messages))
    }
}

func TestImageInTriggeringMessageIsPassedToLLM(t *testing.T) {
    handler, sender, llmClient := newEventTestHandler(t)
    sender.attachments["image-1"] = []byte("image bytes")

    if err := handler.HandleIncomingMessage(models.IncomingMessage{
        GroupID:   "group",
        Author:    "Alice",
        Text:      "@gemini what is this?",
        Timestamp: 1,
        Attachments: []models.Attachment{
            {ID: "image-1", ContentType: "image/png", Size: 11},
        },
    }, true); err != nil {
        t.Fatalf("HandleIncomingMessage failed: %v", err)
    }

    if llmClient.responseCalls != 1 {
        t.Fatalf("response calls = %d, want 1", llmClient.responseCalls)
    }
    if len(llmClient.images) != 1 {
        t.Fatalf("images = %d, want 1", len(llmClient.images))
    }
    if string(llmClient.images[0].Data) != "image bytes" {
        t.Errorf("image data = %q, want %q", string(llmClient.images[0].Data), "image bytes")
    }
}

func TestImageFromQuotedMessageIsPassedToLLM(t *testing.T) {
    handler, sender, llmClient := newEventTestHandler(t)
    sender.attachments["image-2"] = []byte("quoted image bytes")

    if err := handler.HandleIncomingMessage(models.IncomingMessage{
        GroupID:   "group",
        Author:    "Alice",
        Timestamp: 10,
        Attachments: []models.Attachment{
            {ID: "image-2", ContentType: "image/jpeg", Size: 18},
        },
    }, true); err != nil {
        t.Fatalf("HandleIncomingMessage failed: %v", err)
    }

    if err := handler.HandleIncomingMessage(models.IncomingMessage{
        GroupID:   "group",
        Author:    "Alice",
        Text:      "@gemini what is in this image?",
        Timestamp: 11,
        Quote: &models.Quote{
            ID: 10,
        },
    }, true); err != nil {
        t.Fatalf("HandleIncomingMessage failed: %v", err)
    }

    if llmClient.responseCalls != 1 {
        t.Fatalf("response calls = %d, want 1", llmClient.responseCalls)
    }
    if len(llmClient.images) != 1 {
        t.Fatalf("images = %d, want 1", len(llmClient.images))
    }
    if string(llmClient.images[0].Data) != "quoted image bytes" {
        t.Errorf("image data = %q, want %q", string(llmClient.images[0].Data), "quoted image bytes")
    }
}
