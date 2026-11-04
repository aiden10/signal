package handlers

import (
    "path/filepath"
    "testing"

    "signal/db"
)

type recordingSender struct {
    messages []string
}

func (s *recordingSender) SendMessage(text, groupId string) error {
    s.messages = append(s.messages, text)
    return nil
}

func newCommandTestHandler(t *testing.T) (*EventHandler, *recordingSender, *SettingsHandler) {
    t.Helper()

    database, err := db.NewDatabase(filepath.Join(t.TempDir(), "test.db"))
    if err != nil {
        t.Fatalf("failed to open test database: %v", err)
    }
    t.Cleanup(func() {
        _ = database.Close()
    })

    if err := database.Migrate(); err != nil {
        t.Fatalf("failed to migrate test database: %v", err)
    }

    settings := NewSettingsHandler(database)
    sender := &recordingSender{}
    return NewEventHandler(nil, settings, nil, sender, "", ""), sender, settings
}

func TestParseCommand(t *testing.T) {
    tests := []struct {
        text    string
        want    Command
        handled bool
    }{
        {text: "@gemini hello", handled: false},
        {text: "hello", handled: false},
        {text: "/", handled: false},
        {text: "  /help  ", want: Command{Name: "help"}, handled: true},
        {text: "/history 50", want: Command{Name: "history", Args: "50"}, handled: true},
        {text: "/System-Prompt You are a pirate.", want: Command{Name: "system-prompt", Args: "You are a pirate."}, handled: true},
        {text: "/system-prompt\nYou are a pirate.\nAlways say arr.", want: Command{Name: "system-prompt", Args: "You are a pirate.\nAlways say arr."}, handled: true},
    }

    for _, test := range tests {
        got, handled := ParseCommand(test.text)
        if handled != test.handled {
            t.Errorf("ParseCommand(%q) handled = %v, want %v", test.text, handled, test.handled)
            continue
        }
        if got != test.want {
            t.Errorf("ParseCommand(%q) = %#v, want %#v", test.text, got, test.want)
        }
    }
}

func TestHistoryCommandUpdatesSettings(t *testing.T) {
    handler, sender, settings := newCommandTestHandler(t)

    if err := handler.HandleDataMessage("group", "Alice", "/history 50", true); err != nil {
        t.Fatalf("HandleDataMessage failed: %v", err)
    }

    if got := settings.Get("group").HistoryLimit; got != 50 {
        t.Errorf("history limit = %d, want 50", got)
    }
    if len(sender.messages) != 1 {
        t.Fatalf("sent %d messages, want 1", len(sender.messages))
    }
}

func TestSystemPromptCommandUpdatesSettings(t *testing.T) {
    handler, sender, settings := newCommandTestHandler(t)
    prompt := "You are a pirate.\nAlways say arr."

    if err := handler.HandleDataMessage("group", "Alice", "/system-prompt "+prompt, true); err != nil {
        t.Fatalf("HandleDataMessage failed: %v", err)
    }

    if got := settings.Get("group").SystemPrompt; got != prompt {
        t.Errorf("system prompt = %q, want %q", got, prompt)
    }
    if len(sender.messages) != 1 {
        t.Fatalf("sent %d messages, want 1", len(sender.messages))
    }
}

func TestInvalidLimitDoesNotChangeSettings(t *testing.T) {
    handler, sender, settings := newCommandTestHandler(t)

    for _, text := range []string{"/history nope", "/history -1", "/relevant 999"} {
        if err := handler.HandleDataMessage("group", "Alice", text, true); err != nil {
            t.Fatalf("HandleDataMessage failed: %v", err)
        }
    }

    got := settings.Get("group")
    if got.HistoryLimit != DefaultHistoryLimit || got.RelevantLimit != DefaultRelevantLimit {
        t.Errorf("limits changed to history=%d relevant=%d, want defaults", got.HistoryLimit, got.RelevantLimit)
    }
    if len(sender.messages) != 3 {
        t.Fatalf("sent %d messages, want 3", len(sender.messages))
    }
}

func TestResetCommandRestoresDefaults(t *testing.T) {
    handler, _, settings := newCommandTestHandler(t)

    if err := handler.HandleDataMessage("group", "Alice", "/history 5", true); err != nil {
        t.Fatalf("HandleDataMessage failed: %v", err)
    }
    if err := handler.HandleDataMessage("group", "Alice", "/system-prompt Be nice.", true); err != nil {
        t.Fatalf("HandleDataMessage failed: %v", err)
    }
    if err := handler.HandleDataMessage("group", "Alice", "/reset", true); err != nil {
        t.Fatalf("HandleDataMessage failed: %v", err)
    }

    got := settings.Get("group")
    if got != DefaultSettings() {
        t.Errorf("settings = %#v, want %#v", got, DefaultSettings())
    }
}

func TestUnknownCommandIsHandled(t *testing.T) {
    handler, sender, _ := newCommandTestHandler(t)

    if err := handler.HandleDataMessage("group", "Alice", "/does-not-exist", true); err != nil {
        t.Fatalf("HandleDataMessage failed: %v", err)
    }

    if len(sender.messages) != 1 {
        t.Fatalf("sent %d messages, want 1", len(sender.messages))
    }
}
