package handlers

import (
    "fmt"
    "strings"

    "signal/utils"
)

type LLMClient interface {
    GenerateResponse(context []Message, initialPrompt string) string
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
            fmt.Printf("failed sending group message: %v\n", err)
        }
    }
}

func (e *EventHandler) HandleDataMessage(groupId, author, text string) {
    fmt.Printf("Message received")
    e.History.Record(groupId, author, text)
    
    if strings.Contains(strings.ToLower(text), "@gemini") {
        sendingId, err := utils.FindSendingId(groupId)
        if e.targetGroup != "" && sendingId != e.targetGroup { 
            fmt.Printf("Not checking message because it was sent to a non-target group")
            return
        }
        context := e.History.GetContext(groupId)
        fmt.Printf("Generating response for group %s using %d messages of context\n", groupId, len(context))
        response := e.LLM.GenerateResponse(context, text)
        e.SendMessage(sendingId, "Bot", response)
    }
}