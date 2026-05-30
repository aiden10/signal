package models

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/gorilla/websocket"
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

func (c *SignalClient) SendMessage(text, groupId string) error {
    body := map[string]any{
        "message":    text,
        "number":     c.PhoneNumber,
        "recipients": []string{groupId},
    }

    raw, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("marshal body: %w", err)
    }

    url := c.BaseURL + "/v2/send"

    req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))

    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")

    log.Printf("Attempting to send message to URL: %s", url)
    log.Printf("Body: %s", string(raw))

    resp, err := c.HTTPClient.Do(req)
    
    if err != nil {
        return fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return fmt.Errorf("send failed: %s", resp.Status)
    }

    return nil
}