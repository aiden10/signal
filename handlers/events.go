package handlers

import (
    "log"
    "strings"

    "signal/utils"
    "signal/models"
)

type LLMClient interface {
    GenerateResponse(relevantMessages, recentMessages []models.Message, initialPrompt string) string
    GenerateEmbedding(text string) ([]float32, error)
}

type MessageSender interface {
    SendMessage(text, groupId string) error
}

type EventHandler struct {
    History *HistoryHandler
    LLM     LLMClient
    Sender  MessageSender
    targetGroup string
    phone string
}

func NewEventHandler(history *HistoryHandler, llm LLMClient, sender MessageSender, targetGroup, phone string) *EventHandler {
    return &EventHandler{
        History: history,
        LLM:     llm,
        Sender:  sender,
        targetGroup: targetGroup,
        phone: phone,
    }
}

func (e *EventHandler) SendMessage(groupId, author, text string) {
    if author == "Bot" && e.Sender != nil {
        if err := e.Sender.SendMessage(text, groupId); err != nil {
            log.Printf("failed sending group message: %v\n", err)
        }
    }
}

func (e *EventHandler) HandleDataMessage(groupId, author, text string, inTest bool) error {
    log.Printf("Message received")
    
    log.Printf("Generating embedding")
    vector, embeddingErr := e.LLM.GenerateEmbedding(text)
    if embeddingErr != nil {
        log.Printf("failed to generate message embedding: %v\n", embeddingErr)
    }

    e.History.Record(groupId, author, text, vector)

    if strings.Contains(strings.ToLower(text), "@gemini") {
        var sendingId = groupId
        var err error

        if !inTest {
            sendingId, err = utils.FindSendingId(groupId)
        }

        if e.targetGroup != "" && sendingId != e.targetGroup {
            log.Printf("Not checking message because it was sent to a non-target group")
            return nil
        }

        if err != nil {
            log.Printf("Error finding sending id: %v", err)
            return err
        }

        relevant, recent := e.History.GetContext(groupId, vector)

        response := e.LLM.GenerateResponse(relevant, recent, text)
        
        geminiResponseVector, geminiEmbeddingError := e.LLM.GenerateEmbedding(response)
        if geminiEmbeddingError != nil {
            log.Printf("failed to generate message embedding: %v\n", geminiEmbeddingError)
        }

        e.History.Record(groupId, "Bot", response, geminiResponseVector)

        e.SendMessage(sendingId, "Bot", response)
    }
    return nil
}