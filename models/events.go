package models

import "strings"

type SignalEnvelope struct {
    Type        string       `json:"type,omitempty"`
    GroupID     string       `json:"groupId,omitempty"`
    Author      string       `json:"author,omitempty"`
    Text        string       `json:"text,omitempty"`
    DataMessage *DataMessage `json:"dataMessage,omitempty"`
    Receipt     *Receipt     `json:"receipt,omitempty"`
    Envelope    *Envelope    `json:"envelope,omitempty"`
}

type Envelope struct {
    Source         string       `json:"source,omitempty"`
    SourceNumber   string       `json:"sourceNumber,omitempty"`
    DataMessage    *DataMessage `json:"dataMessage,omitempty"`
    ReceiptMessage *Receipt     `json:"receiptMessage,omitempty"`
    SyncMessage    *SyncMessage `json:"syncMessage,omitempty"`
}

type SyncMessage struct {
    SentMessage *SentMessage `json:"sentMessage,omitempty"`
}

type SentMessage struct {
    Destination string       `json:"destination,omitempty"`
    Message     string       `json:"message,omitempty"`
    Timestamp   int64        `json:"timestamp,omitempty"`
    GroupID     string       `json:"groupId,omitempty"`
    GroupInfo   *GroupInfo   `json:"groupInfo,omitempty"`
    Attachments []Attachment `json:"attachments,omitempty"`
    Quote       *Quote       `json:"quote,omitempty"`
}

type DataMessage struct {
    Message     string       `json:"message,omitempty"`
    Text        string       `json:"text,omitempty"`
    Timestamp   int64        `json:"timestamp,omitempty"`
    GroupID     string       `json:"groupId,omitempty"`
    GroupInfo   *GroupInfo   `json:"groupInfo,omitempty"`
    Attachments []Attachment `json:"attachments,omitempty"`
    Quote       *Quote       `json:"quote,omitempty"`
}

type Attachment struct {
    ID          string `json:"id,omitempty"`
    ContentType string `json:"contentType,omitempty"`
    Filename    string `json:"filename,omitempty"`
    Caption     string `json:"caption,omitempty"`
    Size        int64  `json:"size,omitempty"`
    Width       int    `json:"width,omitempty"`
    Height      int    `json:"height,omitempty"`
}

type Quote struct {
    ID           int64              `json:"id"`
    Author       string             `json:"author,omitempty"`
    AuthorNumber string             `json:"authorNumber,omitempty"`
    AuthorUUID   string             `json:"authorUuid,omitempty"`
    Text         string             `json:"text,omitempty"`
    Attachments  []QuotedAttachment `json:"attachments,omitempty"`
}

type QuotedAttachment struct {
    ContentType string      `json:"contentType,omitempty"`
    Filename    string      `json:"filename,omitempty"`
    Thumbnail   *Attachment `json:"thumbnail,omitempty"`
}

type Message struct {
    Author string
    Text   string
}

type Image struct {
    ContentType string
    Data        []byte
}

type IncomingMessage struct {
    GroupID     string
    Author      string
    Text        string
    Timestamp   int64
    Attachments []Attachment
    Quote       *Quote
    IsData      bool
    IsReceipt   bool
}

type GroupInfo struct {
    GroupID string `json:"groupId,omitempty"`
}

type Receipt struct {
    When int64 `json:"when,omitempty"`
}

func (s SignalEnvelope) NormalizedMessage() IncomingMessage {
    message := IncomingMessage{
        GroupID: s.GroupID,
        Author:  s.Author,
        Text:    s.Text,
    }

    if s.DataMessage != nil {
        message.applyDataMessage(s.DataMessage)
    }
    if s.Receipt != nil {
        message.IsReceipt = true
    }

    if s.Envelope != nil {
        if message.Author == "" {
            message.Author = s.Envelope.Source
        }
        if message.Author == "" {
            message.Author = s.Envelope.SourceNumber
        }

        if s.Envelope.DataMessage != nil {
            message.applyDataMessage(s.Envelope.DataMessage)
        }

        if s.Envelope.SyncMessage != nil && s.Envelope.SyncMessage.SentMessage != nil {
            sent := s.Envelope.SyncMessage.SentMessage
            message.IsData = true
            if message.Text == "" {
                message.Text = sent.Message
            }
            if message.Text == "" {
                message.Text = attachmentText(sent.Attachments)
            }
            if message.GroupID == "" {
                if sent.GroupID != "" {
                    message.GroupID = sent.GroupID
                } else if sent.GroupInfo != nil {
                    message.GroupID = sent.GroupInfo.GroupID
                }
            }
            if message.Timestamp == 0 {
                message.Timestamp = sent.Timestamp
            }
            if len(message.Attachments) == 0 {
                message.Attachments = sent.Attachments
            }
            if message.Quote == nil {
                message.Quote = sent.Quote
            }
        }

        if s.Envelope.ReceiptMessage != nil {
            message.IsReceipt = true
        }
    }

    return message
}

func (s SignalEnvelope) Normalized() (groupID, author, text string, isData, isReceipt bool) {
    message := s.NormalizedMessage()
    return message.GroupID, message.Author, message.Text, message.IsData, message.IsReceipt
}

func (m *IncomingMessage) applyDataMessage(data *DataMessage) {
    m.IsData = true
    if m.Text == "" {
        if data.Text != "" {
            m.Text = data.Text
        } else {
            m.Text = data.Message
        }
    }
    if m.Text == "" {
        m.Text = attachmentText(data.Attachments)
    }
    if m.GroupID == "" {
        if data.GroupID != "" {
            m.GroupID = data.GroupID
        } else if data.GroupInfo != nil {
            m.GroupID = data.GroupInfo.GroupID
        }
    }
    if m.Timestamp == 0 {
        m.Timestamp = data.Timestamp
    }
    if len(m.Attachments) == 0 {
        m.Attachments = data.Attachments
    }
    if m.Quote == nil {
        m.Quote = data.Quote
    }
}

func attachmentText(attachments []Attachment) string {
    var captions []string
    for _, attachment := range attachments {
        caption := strings.TrimSpace(attachment.Caption)
        if caption != "" {
            captions = append(captions, caption)
        }
    }
    return strings.Join(captions, "\n")
}
