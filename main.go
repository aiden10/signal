package main

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "signal/handlers"
)

type SignalClient struct {
    BaseURL     string
    PhoneNumber string
    HTTPClient  *http.Client
    Logger      *log.Logger
}

func NewSignalClient(baseURL, phoneNumber string, logger *log.Logger) *SignalClient {
    return &SignalClient{
        BaseURL:     strings.TrimRight(baseURL, "/"),
        PhoneNumber: phoneNumber,
        HTTPClient: &http.Client{
            Timeout: 15 * time.Second,
        },
        Logger: logger,
    }
}

func (c *SignalClient) SendGroupMessage(groupID, text string) error {
    body := map[string]any{
        "message": text,
        "number":  c.PhoneNumber,
    }

    if strings.HasPrefix(groupID, "group.") {
        body["groupId"] = groupID
    } else if groupID != "" {
        body["recipients"] = []string{groupID}
    } else {
        // If we messaged ourselves and groupID is empty, send back to ourselves
        body["recipients"] = []string{c.PhoneNumber}
    }

    raw, err := json.Marshal(body)
    if err != nil {
        return err
    }

    req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v2/send", bytes.NewReader(raw))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        c.Logger.Printf("send failed with status: %s", resp.Status)
    }

    return nil
}

func loadDotEnv(path string) error {
    raw, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    lines := strings.Split(string(raw), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
            continue
        }

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key := strings.TrimSpace(parts[0])
        val := strings.TrimSpace(parts[1])

        // keep existing shell-provided vars
        if os.Getenv(key) == "" {
            _ = os.Setenv(key, val)
        }
    }

    return nil
}

func main() {
    logger := log.Default()
    _ = loadDotEnv(".env")

    serverURL := os.Getenv("SERVER_URL")
    socketURL := os.Getenv("SOCKET_URL")
    phone := os.Getenv("PHONE_NUMBER")
    targetGroup := os.Getenv("GROUP_ID")

    if socketURL == "" {
        logger.Fatal("SOCKET_URL is required")
    }
    if serverURL == "" {
        logger.Fatal("SERVER_URL is required")
    }

    history := handlers.NewHistoryHandler(20)
    llm := NewLLMProvider(os.Getenv("GEMINI_API_KEY"), os.Getenv("GEMINI_ENDPOINT"))
    sender := NewSignalClient(serverURL, phone, logger)

    eventHandler := handlers.NewEventHandler(history, llm, sender, targetGroup)
    socketHandler := handlers.NewSocketHandler(socketURL, eventHandler, logger)

    logger.Println("starting socket handler...")
    socketHandler.Start()
}