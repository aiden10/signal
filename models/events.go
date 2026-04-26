package models

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
    GroupID     string       `json:"groupId,omitempty"`
    GroupInfo   *GroupInfo   `json:"groupInfo,omitempty"`
}

type DataMessage struct {
    Message   string     `json:"message,omitempty"`
    Text      string     `json:"text,omitempty"`
    GroupID   string     `json:"groupId,omitempty"`
    GroupInfo *GroupInfo `json:"groupInfo,omitempty"`
}

type GroupInfo struct {
    GroupID string `json:"groupId,omitempty"`
}

type Receipt struct {
    When int64 `json:"when,omitempty"`
}

func (s SignalEnvelope) Normalized() (groupID, author, text string, isData, isReceipt bool) {
    groupID = s.GroupID
    author = s.Author
    text = s.Text

    if s.DataMessage != nil {
        isData = true
        if text == "" {
            if s.DataMessage.Text != "" {
                text = s.DataMessage.Text
            } else {
                text = s.DataMessage.Message
            }
        }
        if groupID == "" {
            if s.DataMessage.GroupID != "" {
                groupID = s.DataMessage.GroupID
            } else if s.DataMessage.GroupInfo != nil {
                groupID = s.DataMessage.GroupInfo.GroupID
            }
        }
    }
    if s.Receipt != nil {
        isReceipt = true
    }

    if s.Envelope != nil {
        if author == "" {
            author = s.Envelope.Source
        }
        
        // Handle standard DataMessages
        if s.Envelope.DataMessage != nil {
            isData = true
            if text == "" {
                if s.Envelope.DataMessage.Text != "" {
                    text = s.Envelope.DataMessage.Text
                } else {
                    text = s.Envelope.DataMessage.Message
                }
            }
            if groupID == "" {
                if s.Envelope.DataMessage.GroupID != "" {
                    groupID = s.Envelope.DataMessage.GroupID
                } else if s.Envelope.DataMessage.GroupInfo != nil {
                    groupID = s.Envelope.DataMessage.GroupInfo.GroupID
                }
            }
        }

        // Handle SyncMessages (messages sent from another device)
        if s.Envelope.SyncMessage != nil && s.Envelope.SyncMessage.SentMessage != nil {
            isData = true
            sent := s.Envelope.SyncMessage.SentMessage
            if text == "" {
                text = sent.Message
            }
            if groupID == "" {
                if sent.GroupID != "" {
                    groupID = sent.GroupID
                } else if sent.GroupInfo != nil {
                    groupID = sent.GroupInfo.GroupID
                }
            }
        }

        if s.Envelope.ReceiptMessage != nil {
            isReceipt = true
        }
    }

    return
}