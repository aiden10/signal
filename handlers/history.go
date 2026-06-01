package handlers

import (
	"log"
	"math"
	"sort"
	
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

func (h *HistoryHandler) GetContext(groupId string, queryVec []float32) ([]models.Message, []models.Message) {
    memories, err := h.database.GetAllMemories(groupId)
    if err != nil {
        log.Printf("Error getting memories: %v", err)
        memories = []db.Memory{}
    }

    recent := h.database.GetRecentMessages(groupId, 20)

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
    for i := 0; i < 10 && i < len(results); i++ {
        relevant = append(relevant, results[i].msg)
    }

    return relevant, recent
}