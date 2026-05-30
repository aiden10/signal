package handlers

import (
    "log"
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
            log.Printf("failed sending group message: %v\n", err)
        }
    }
}

func (e *EventHandler) HandleDataMessage(groupId, author, text string, inTest bool) error {
    log.Printf("Message received")
    e.History.Record(groupId, author, text)
    
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

        context := e.History.GetContext(groupId)
        log.Printf("Generating response for group %s using %d messages of context\n", groupId, len(context))
        response := e.LLM.GenerateResponse(context, text)
        e.SendMessage(sendingId, "Bot", response)
    }
    return nil
}