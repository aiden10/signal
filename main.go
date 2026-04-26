package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "signal/handlers"
)

type SignalClient struct {
    BaseURL     string
    PhoneNumber string
    TargetGroup string
    Logger      *log.Logger
    HTTPClient  *http.Client
    
    mu     sync.Mutex
    wsConn *websocket.Conn
}

func NewSignalClient(baseURL, phoneNumber, targetGroup string, logger *log.Logger) *SignalClient {
    return &SignalClient{
        BaseURL:     strings.TrimRight(baseURL, "/"),
        PhoneNumber: phoneNumber,
        TargetGroup: targetGroup,
        Logger:      logger,
        HTTPClient:  &http.Client{Timeout: 15 * time.Second},
    }
}

func (c *SignalClient) SendGroupMessage(text string) error {
    body := map[string]any{
        "message":    text,
        "number":     c.PhoneNumber,
        "recipients": []string{c.TargetGroup},
    }

    raw, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("marshal body: %w", err)
    }

    c.Logger.Printf("sending to /v2/send: %s", raw)

    req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v2/send", bytes.NewReader(raw))
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        buf := new(bytes.Buffer)
        buf.ReadFrom(resp.Body)
        return fmt.Errorf("send failed (%s): %s", resp.Status, buf.String())
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
    sender := NewSignalClient(serverURL, phone, targetGroup, logger)

    eventHandler := handlers.NewEventHandler(history, llm, sender, targetGroup)
    socketHandler := handlers.NewSocketHandler(socketURL, eventHandler, logger)

    logger.Println("starting socket handler...")
    socketHandler.Start()
}