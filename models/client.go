package models

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
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

func (c *SignalClient) SendMessage(text, groupId string) (string, error) {
    body := map[string]any{
        "message":    text,
        "number":     c.PhoneNumber,
        "recipients": []string{groupId},
    }

    raw, err := json.Marshal(body)
    if err != nil {
        return "", fmt.Errorf("marshal body: %w", err)
    }

    endpoint := c.BaseURL + "/v2/send"

    req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))

    if err != nil {
        return "", fmt.Errorf("create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")

    log.Printf("Attempting to send message to URL: %s", endpoint)
    log.Printf("Body: %s", string(raw))

    resp, err := c.HTTPClient.Do(req)
    
    if err != nil {
        return "", fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return "", fmt.Errorf("send failed: %s", resp.Status)
    }

    var result struct {
        Timestamp string `json:"timestamp"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        log.Printf("failed to decode send response: %v", err)
        return "", nil
    }

    return result.Timestamp, nil
}

func (c *SignalClient) GetAttachment(id string) ([]byte, error) {
    if strings.TrimSpace(id) == "" {
        return nil, fmt.Errorf("attachment id is empty")
    }

    endpoint := c.BaseURL + "/v1/attachments/" + url.PathEscape(id)
    resp, err := c.HTTPClient.Get(endpoint)
    if err != nil {
        return nil, fmt.Errorf("get attachment: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return nil, fmt.Errorf("get attachment failed: %s", resp.Status)
    }

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read attachment: %w", err)
    }
    return data, nil
}
