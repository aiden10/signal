package handlers

import (
    "log"
    "strconv"

    "signal/db"
)

const (
    DefaultHistoryLimit  = 20
    DefaultRelevantLimit = 10
    DefaultSystemPrompt  = "You are responding in a group chat. Keep responses under 2000 characters and sound more human. Do not include any markdown formatting and occasionally use emojis and slang."

    settingHistoryLimit  = "history_limit"
    settingRelevantLimit = "relevant_limit"
    settingSystemPrompt  = "system_prompt"
)

type Settings struct {
    HistoryLimit  int
    RelevantLimit int
    SystemPrompt  string
}

type SettingsHandler struct {
    database *db.Database
}

func NewSettingsHandler(database *db.Database) *SettingsHandler {
    return &SettingsHandler{
        database: database,
    }
}

func DefaultSettings() Settings {
    return Settings{
        HistoryLimit:  DefaultHistoryLimit,
        RelevantLimit: DefaultRelevantLimit,
        SystemPrompt:  DefaultSystemPrompt,
    }
}

func (h *SettingsHandler) Get(groupId string) Settings {
    settings := DefaultSettings()

    values, err := h.database.GetSettings(groupId)
    if err != nil {
        log.Printf("Error getting settings: %v", err)
        return settings
    }

    settings.HistoryLimit = intSetting(values, settingHistoryLimit, settings.HistoryLimit)
    settings.RelevantLimit = intSetting(values, settingRelevantLimit, settings.RelevantLimit)
    if value, ok := values[settingSystemPrompt]; ok && value != "" {
        settings.SystemPrompt = value
    }

    return settings
}

func (h *SettingsHandler) SetHistoryLimit(groupId string, limit int) error {
    return h.database.SetSetting(groupId, settingHistoryLimit, strconv.Itoa(limit))
}

func (h *SettingsHandler) SetRelevantLimit(groupId string, limit int) error {
    return h.database.SetSetting(groupId, settingRelevantLimit, strconv.Itoa(limit))
}

func (h *SettingsHandler) SetSystemPrompt(groupId, prompt string) error {
    return h.database.SetSetting(groupId, settingSystemPrompt, prompt)
}

func (h *SettingsHandler) Reset(groupId string) error {
    return h.database.DeleteSettings(groupId)
}

func intSetting(values map[string]string, key string, fallback int) int {
    raw, ok := values[key]
    if !ok {
        return fallback
    }

    value, err := strconv.Atoi(raw)
    if err != nil {
        return fallback
    }
    return value
}
