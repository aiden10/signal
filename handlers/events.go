package handlers

import (
    "fmt"
    "log"
    "strings"

    "signal/models"
    "signal/utils"
)

type LLMClient interface {
    GenerateResponse(relevantMessages, recentMessages []models.Message, images []models.Image, systemPrompt, initialPrompt string) string
    GenerateEmbedding(text string) ([]float32, error)
}

type MessageSender interface {
    SendMessage(text, groupId string) (string, error)
}

type AttachmentFetcher interface {
    GetAttachment(id string) ([]byte, error)
}

type EventHandler struct {
    History     *HistoryHandler
    Settings    *SettingsHandler
    Commands    *CommandHandler
    LLM         LLMClient
    Sender      MessageSender
    Attachments AttachmentFetcher
    targetGroup string
    phone       string
}

func NewEventHandler(history *HistoryHandler, settings *SettingsHandler, llm LLMClient, sender MessageSender, targetGroup, phone string) *EventHandler {
    handler := &EventHandler{
        History:     history,
        Settings:    settings,
        Commands:    NewCommandHandler(settings),
        LLM:         llm,
        Sender:      sender,
        targetGroup: targetGroup,
        phone:       phone,
    }

    if fetcher, ok := sender.(AttachmentFetcher); ok {
        handler.Attachments = fetcher
    }

    return handler
}

func (e *EventHandler) SendBotMessage(groupId, sendingId, text string) {
    if e.Sender == nil {
        return
    }

    timestamp, err := e.Sender.SendMessage(text, sendingId)
    if err != nil {
        log.Printf("failed sending group message: %v\n", err)
        return
    }
    if strings.TrimSpace(timestamp) == "" {
        log.Printf("sent message without a timestamp; replies to it will not trigger the bot")
        return
    }
    if e.History != nil {
        if err := e.History.RecordBotMessage(groupId, timestamp); err != nil {
            log.Printf("failed recording bot message timestamp: %v", err)
        }
    }
}

func (e *EventHandler) HandleDataMessage(groupId, author, text string, inTest bool) error {
    return e.HandleIncomingMessage(models.IncomingMessage{
        GroupID: groupId,
        Author:  author,
        Text:    text,
    }, inTest)
}

func (e *EventHandler) HandleIncomingMessage(message models.IncomingMessage, inTest bool) error {
    log.Printf("Message received")

    groupId := message.GroupID
    sendingId := groupId
    var err error
    if !inTest {
        sendingId, err = utils.FindSendingId(groupId)
    }

    if e.targetGroup != "" && sendingId != e.targetGroup {
        log.Printf("Not checking message because it was sent to a non-target group")
        return nil
    }

    if e.History != nil && message.Timestamp != 0 && len(message.Attachments) > 0 {
        if err := e.History.RecordAttachments(groupId, message.Timestamp, message.Attachments); err != nil {
            log.Printf("failed recording message attachments: %v", err)
        }
    }

    text := message.Text
    if command, ok := ParseCommand(text); ok {
        if err != nil {
            log.Printf("Error finding sending id: %v", err)
            return err
        }

        response, _ := e.Commands.Handle(groupId, command)
        log.Printf("Handled command /%s", command.Name)
        e.SendBotMessage(groupId, sendingId, response)
        return nil
    }

    replyToBot := false
    if message.Quote != nil && e.History != nil {
        replyToBot = e.History.IsBotMessage(groupId, message.Quote.ID)
    }
    if replyToBot {
        log.Printf("Message is a reply to a bot message")
    }

    var vector []float32
    log.Printf("Generating embedding")
    if e.History != nil && strings.TrimSpace(text) != "" {
        var embeddingErr error
        vector, embeddingErr = e.LLM.GenerateEmbedding(text)
        if embeddingErr != nil {
            log.Printf("failed to generate embedding: %v", embeddingErr)
        } else {
            e.History.Record(groupId, message.Author, text, vector)
        }
    }

    if strings.Contains(strings.ToLower(text), "@gemini") || replyToBot {
        if err != nil {
            log.Printf("Error finding sending id: %v", err)
            return err
        }

        settings := e.Settings.Get(groupId)
        relevant, recent := e.History.GetContext(groupId, vector, settings.HistoryLimit, settings.RelevantLimit)
        images := e.imagesForMessage(groupId, message)
        prompt := buildUserRequest(text, message.Quote, replyToBot, len(images) > 0)
        log.Printf("Generating response with %d relevant messages, %d recent messages, and %d images", len(relevant), len(recent), len(images))
        response := e.LLM.GenerateResponse(relevant, recent, images, settings.SystemPrompt, prompt)

        geminiResponseVector, geminiEmbeddingError := e.LLM.GenerateEmbedding(response)
        if geminiEmbeddingError != nil {
            log.Printf("failed to generate message embedding: %v\n", geminiEmbeddingError)
        }
        e.History.Record(groupId, "Bot", response, geminiResponseVector)
        e.SendBotMessage(groupId, sendingId, response)
    }
    return nil
}

func buildUserRequest(text string, quote *models.Quote, replyToBot, hasImages bool) string {
    text = strings.TrimSpace(text)
    parts := []string{}

    if quote != nil {
        quotedText := strings.TrimSpace(quote.Text)
        if quotedText != "" {
            label := "Quoted message"
            if replyToBot {
                label = "Your earlier message"
            }
            parts = append(parts, fmt.Sprintf("%s: %s", label, quotedText))
        }
    }

    if text != "" {
        parts = append(parts, text)
    } else if hasImages {
        if replyToBot {
            parts = append(parts, "The user replied to your message with this image.")
        } else {
            parts = append(parts, "The user sent this image.")
        }
    } else if replyToBot {
        parts = append(parts, "The user replied to your message.")
    }

    return strings.Join(parts, "\n\n")
}
