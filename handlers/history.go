package handlers

import (
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	
	"signal/db"
	"signal/models"
)

type HistoryHandler struct {
	database *db.Database
}

func NewHistoryHandler(database *db.Database) *HistoryHandler {
	return &HistoryHandler {
		database: database,
	}
}

func (h *HistoryHandler) Record(groupId, author, text string, vector []float32) {
	h.database.InsertMemory(groupId, author, text, vector)
}

func (h *HistoryHandler) RecordBotMessage(groupId, timestamp string) error {
    parsed, err := strconv.ParseInt(strings.TrimSpace(timestamp), 10, 64)
    if err != nil {
        return err
    }
    return h.database.InsertBotMessage(groupId, parsed)
}

func (h *HistoryHandler) IsBotMessage(groupId string, timestamp int64) bool {
    isBotMessage, err := h.database.IsBotMessage(groupId, timestamp)
    if err != nil {
        log.Printf("Error checking bot message: %v", err)
        return false
    }
    return isBotMessage
}

func (h *HistoryHandler) RecordAttachments(groupId string, timestamp int64, attachments []models.Attachment) error {
    return h.database.InsertMessageAttachments(groupId, timestamp, attachments)
}

func (h *HistoryHandler) GetAttachments(groupId string, timestamp int64) []models.Attachment {
    attachments, err := h.database.GetMessageAttachments(groupId, timestamp)
    if err != nil {
        log.Printf("Error getting message attachments: %v", err)
        return nil
    }
    return attachments
}

func cosineSimilarity(a, b []float32) float64 {
    var dot, normA, normB float64
    for i := range a {
        dot += float64(a[i]) * float64(b[i])
        normA += float64(a[i]) * float64(a[i])
        normB += float64(b[i]) * float64(b[i])
    }

    if normA == 0 || normB == 0 {
        return 0
    }
    return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (h *HistoryHandler) GetContext(groupId string, queryVec []float32, recentLimit, relevantLimit int) ([]models.Message, []models.Message) {
    var recent []models.Message
    if recentLimit > 0 {
        recent = h.database.GetRecentMessages(groupId, recentLimit)
    }

    if relevantLimit <= 0 {
        return nil, recent
    }

    memories, err := h.database.GetAllMemories(groupId)
    if err != nil {
        log.Printf("Error getting memories: %v", err)
        memories = []db.Memory{}
    }

    type scored struct {
        msg   models.Message
        score float64
    }

    var results []scored
    for _, m := range memories {
        if len(m.Vector) == 0 || len(m.Vector) != len(queryVec) {
            log.Println("Skipping cosine similarity because vector has no related memories")
            continue // skip malformed entries
        }
        
        score := cosineSimilarity(queryVec, m.Vector)
        results = append(results, scored{
            msg:   models.Message{Author: m.Author, Text: m.Content},
            score: score,
        })
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].score > results[j].score
    })

    var relevant []models.Message
    for i := 0; i < relevantLimit && i < len(results); i++ {
        relevant = append(relevant, results[i].msg)
    }

    return relevant, recent
}