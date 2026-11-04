package models

import "testing"

func TestNormalizedMessageIncludesQuoteAndAttachments(t *testing.T) {
    envelope := SignalEnvelope{
        Envelope: &Envelope{
            Source: "+15550001111",
            DataMessage: &DataMessage{
                Message:   "what is this?",
                Timestamp: 100,
                GroupInfo: &GroupInfo{GroupID: "group"},
                Attachments: []Attachment{
                    {ID: "image.jpg", ContentType: "image/jpeg"},
                },
                Quote: &Quote{
                    ID:   99,
                    Text: "previous message",
                },
            },
        },
    }

    message := envelope.NormalizedMessage()
    if !message.IsData {
        t.Fatal("IsData = false, want true")
    }
    if message.GroupID != "group" {
        t.Errorf("GroupID = %q, want %q", message.GroupID, "group")
    }
    if message.Author != "+15550001111" {
        t.Errorf("Author = %q, want %q", message.Author, "+15550001111")
    }
    if message.Text != "what is this?" {
        t.Errorf("Text = %q, want %q", message.Text, "what is this?")
    }
    if message.Timestamp != 100 {
        t.Errorf("Timestamp = %d, want 100", message.Timestamp)
    }
    if len(message.Attachments) != 1 || message.Attachments[0].ID != "image.jpg" {
        t.Errorf("Attachments = %#v", message.Attachments)
    }
    if message.Quote == nil || message.Quote.ID != 99 {
        t.Errorf("Quote = %#v", message.Quote)
    }
}

func TestNormalizedMessageIncludesSyncMessageData(t *testing.T) {
    envelope := SignalEnvelope{
        Envelope: &Envelope{
            Source: "+15550001111",
            SyncMessage: &SyncMessage{
                SentMessage: &SentMessage{
                    Message:   "my image",
                    Timestamp: 200,
                    GroupID:   "group",
                    Attachments: []Attachment{
                        {ID: "image.png", ContentType: "image/png"},
                    },
                    Quote: &Quote{ID: 199},
                },
            },
        },
    }

    message := envelope.NormalizedMessage()
    if !message.IsData {
        t.Fatal("IsData = false, want true")
    }
    if message.Text != "my image" {
        t.Errorf("Text = %q, want %q", message.Text, "my image")
    }
    if message.Timestamp != 200 {
        t.Errorf("Timestamp = %d, want 200", message.Timestamp)
    }
    if len(message.Attachments) != 1 || message.Attachments[0].ID != "image.png" {
        t.Errorf("Attachments = %#v", message.Attachments)
    }
    if message.Quote == nil || message.Quote.ID != 199 {
        t.Errorf("Quote = %#v", message.Quote)
    }
}

func TestNormalizedMessageUsesAttachmentCaption(t *testing.T) {
    envelope := SignalEnvelope{
        Envelope: &Envelope{
            Source: "+15550001111",
            DataMessage: &DataMessage{
                Timestamp: 300,
                GroupInfo: &GroupInfo{GroupID: "group"},
                Attachments: []Attachment{
                    {ID: "image.jpg", ContentType: "image/jpeg", Caption: "@gemini describe this"},
                },
            },
        },
    }

    message := envelope.NormalizedMessage()
    if message.Text != "@gemini describe this" {
        t.Errorf("Text = %q, want attachment caption", message.Text)
    }
}
