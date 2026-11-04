package handlers

import (
    "fmt"
    "log"
    "strconv"
    "strings"
    "unicode"
)

const MaxContextMessages = 200

type Command struct {
    Name string
    Args string
}

type CommandHandler struct {
    settings *SettingsHandler
}

func NewCommandHandler(settings *SettingsHandler) *CommandHandler {
    return &CommandHandler{
        settings: settings,
    }
}

func ParseCommand(text string) (Command, bool) {
    trimmed := strings.TrimSpace(text)
    if !strings.HasPrefix(trimmed, "/") {
        return Command{}, false
    }

    body := strings.TrimSpace(strings.TrimPrefix(trimmed, "/"))
    if body == "" {
        return Command{}, false
    }

    index := strings.IndexFunc(body, unicode.IsSpace)
    if index == -1 {
        return Command{Name: strings.ToLower(body)}, true
    }

    return Command{
        Name: strings.ToLower(body[:index]),
        Args: strings.TrimSpace(body[index:]),
    }, true
}

func (h *CommandHandler) Handle(groupId string, command Command) (string, bool) {
    switch command.Name {
    case "help":
        return h.helpText(groupId), true
    case "history":
        return h.updateHistoryLimit(groupId, command.Args), true
    case "relevant":
        return h.updateRelevantLimit(groupId, command.Args), true
    case "system-prompt":
        return h.updateSystemPrompt(groupId, command.Args), true
    case "settings":
        return h.settingsText(groupId), true
    case "reset":
        return h.reset(groupId), true
    case "ping":
        return "Pong! 🏓", true
    default:
        return fmt.Sprintf("Unknown command /%s. Use /help to see available commands.", command.Name), true
    }
}

func (h *CommandHandler) updateHistoryLimit(groupId, args string) string {
    if args == "" {
        return fmt.Sprintf(
            "History context is %d recent messages. Use /history <0-%d> to change it.",
            h.settings.Get(groupId).HistoryLimit,
            MaxContextMessages,
        )
    }

    limit, ok := parseLimit(args)
    if !ok {
        return fmt.Sprintf("Usage: /history <0-%d>", MaxContextMessages)
    }

    if err := h.settings.SetHistoryLimit(groupId, limit); err != nil {
        log.Printf("Failed to update history limit: %v", err)
        return "Failed to update the history context."
    }

    return fmt.Sprintf("History context set to %d recent messages.", limit)
}

func (h *CommandHandler) updateRelevantLimit(groupId, args string) string {
    if args == "" {
        return fmt.Sprintf(
            "Relevant context is %d similar messages. Use /relevant <0-%d> to change it.",
            h.settings.Get(groupId).RelevantLimit,
            MaxContextMessages,
        )
    }

    limit, ok := parseLimit(args)
    if !ok {
        return fmt.Sprintf("Usage: /relevant <0-%d>", MaxContextMessages)
    }

    if err := h.settings.SetRelevantLimit(groupId, limit); err != nil {
        log.Printf("Failed to update relevant limit: %v", err)
        return "Failed to update the relevant context."
    }

    return fmt.Sprintf("Relevant context set to %d similar messages.", limit)
}

func (h *CommandHandler) updateSystemPrompt(groupId, args string) string {
    if args == "" {
        return "Current system prompt:\n" + h.settings.Get(groupId).SystemPrompt
    }

    if err := h.settings.SetSystemPrompt(groupId, args); err != nil {
        log.Printf("Failed to update system prompt: %v", err)
        return "Failed to update the system prompt."
    }

    return "System prompt updated. Use /system-prompt without arguments to view it."
}

func (h *CommandHandler) settingsText(groupId string) string {
    settings := h.settings.Get(groupId)
    return fmt.Sprintf(
        "Settings:\nHistory limit: %d\nRelevant limit: %d\nSystem prompt: %s",
        settings.HistoryLimit,
        settings.RelevantLimit,
        summarize(settings.SystemPrompt, 200),
    )
}

func (h *CommandHandler) helpText(groupId string) string {
    settings := h.settings.Get(groupId)
    return fmt.Sprintf(`Available commands:
/history <n> - recent messages sent to the AI (current: %d)
/relevant <n> - similar messages sent to the AI (current: %d)
/system-prompt <prompt> - set the system prompt, or view it without arguments
/settings - show the current settings
/reset - restore the default settings
/ping - check whether the bot is online

Anyone in the chat can use these commands.`, settings.HistoryLimit, settings.RelevantLimit)
}

func (h *CommandHandler) reset(groupId string) string {
    if err := h.settings.Reset(groupId); err != nil {
        log.Printf("Failed to reset settings: %v", err)
        return "Failed to reset the settings."
    }
    return "Settings reset to defaults."
}

func parseLimit(args string) (int, bool) {
    limit, err := strconv.Atoi(args)
    if err != nil || limit < 0 || limit > MaxContextMessages {
        return 0, false
    }
    return limit, true
}

func summarize(text string, max int) string {
    text = strings.Join(strings.Fields(text), " ")
    runes := []rune(text)
    if len(runes) <= max {
        return text
    }
    return string(runes[:max]) + "..."
}
