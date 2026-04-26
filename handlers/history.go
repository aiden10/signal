package handlers

import "sync"

type Message struct {
	Author string
	Text string
}

type HistoryHandler struct {
	mu sync.RWMutex
	Messages map[string][]Message
	MessageLimit int
}

func NewHistoryHandler(limit int) *HistoryHandler {
	return &HistoryHandler {
		Messages: make(map[string][]Message),
		MessageLimit: limit,
	}
}

func (h *HistoryHandler) Record(groupId, author, text string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Messages[groupId] = append(h.Messages[groupId], Message{
		Author: author, 
		Text: text,
	})
	
	if len(h.Messages[groupId]) > h.MessageLimit {
		h.Messages[groupId] = h.Messages[groupId][1:]
	}
}

func (h *HistoryHandler) GetContext(groupId string) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.Messages[groupId]
}