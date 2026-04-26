package handlers

import (
    "fmt"
    "strings"
)

type LLMClient interface {
    GenerateResponse(context []Message, initialPrompt string) string
}

type MessageSender interface {
    SendGroupMessage(groupID, text string) error
}

type EventHandler struct {
    History *HistoryHandler
    LLM     LLMClient
    Sender  MessageSender
    targetGroup string
}

func NewEventHandler(history *HistoryHandler, llm LLMClient, sender MessageSender, targetGroup string) *EventHandler {
    return &EventHandler{
        History: history,
        LLM:     llm,
        Sender:  sender,
        targetGroup: targetGroup,
    }
}

func (e *EventHandler) SendMessage(groupId, author, text string) {
    e.History.Record(groupId, author, text)

    if author == "LLM" && e.Sender != nil {
        if err := e.Sender.SendGroupMessage(groupId, text); err != nil {
            fmt.Printf("failed sending group message: %v\n", err)
        }
    }
}

func (e *EventHandler) HandleDataMessage(groupId, author, text string) {
    fmt.Printf("Message received")
    if e.targetGroup != "" && groupId != e.targetGroup { 
        fmt.Printf("Not checking message because it was sent to a non-target group")
        return
    }
    e.History.Record(groupId, author, text)

    if strings.Contains(strings.ToLower(text), "@gemini") {
        context := e.History.GetContext(groupId)
        fmt.Printf("Generating response for group %s using %d messages of context\n", groupId, len(context))
        response := e.LLM.GenerateResponse(context, text)
        e.SendMessage(groupId, "LLM", response)
    }
}