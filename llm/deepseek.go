package llm

import (
    "bytes"
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "signal/models"
)

const (
    DefaultDeepSeekBaseURL = "https://api.deepseek.com"
    DefaultDeepSeekModel   = "deepseek-flash"
)

type DeepSeekClient struct {
    apiKey     string
    model      string
    baseURL    string
    httpClient *http.Client
}

type deepSeekRequestMessage struct {
    Role    string `json:"role"`
    Content any    `json:"content"`
}

type deepSeekResponseMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type deepSeekContentPart struct {
    Type     string            `json:"type"`
    Text     string            `json:"text,omitempty"`
    ImageURL *deepSeekImageURL `json:"image_url,omitempty"`
}

type deepSeekImageURL struct {
    URL string `json:"url"`
}

type deepSeekRequest struct {
    Model    string                  `json:"model"`
    Messages []deepSeekRequestMessage `json:"messages"`
    Stream   bool                    `json:"stream"`
}

type deepSeekResponse struct {
    Choices []struct {
        Message deepSeekResponseMessage `json:"message"`
    } `json:"choices"`
}

func NewDeepSeekClient(apiKey, model, baseURL string) *DeepSeekClient {
    model = strings.TrimSpace(model)
    if model == "" {
        model = DefaultDeepSeekModel
    }

    baseURL = strings.TrimSpace(baseURL)
    if baseURL == "" {
        baseURL = DefaultDeepSeekBaseURL
    }

    return &DeepSeekClient{
        apiKey:     strings.TrimSpace(apiKey),
        model:      model,
        baseURL:    strings.TrimRight(baseURL, "/"),
        httpClient: &http.Client{Timeout: 60 * time.Second},
    }
}

func (c *DeepSeekClient) GenerateResponse(ctx context.Context, systemPrompt, prompt string, images []models.Image) (string, error) {
    messages := make([]deepSeekRequestMessage, 0, 2)
    if strings.TrimSpace(systemPrompt) != "" {
        messages = append(messages, deepSeekRequestMessage{
            Role:    "system",
            Content: systemPrompt,
        })
    }

    var content any = prompt
    if len(images) > 0 {
        parts := make([]deepSeekContentPart, 0, len(images)+1)
        parts = append(parts, deepSeekContentPart{
            Type: "text",
            Text: prompt,
        })
        for _, image := range images {
            contentType := strings.TrimSpace(image.ContentType)
            if contentType == "" {
                contentType = http.DetectContentType(image.Data)
            }
            parts = append(parts, deepSeekContentPart{
                Type: "image_url",
                ImageURL: &deepSeekImageURL{
                    URL: fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(image.Data)),
                },
            })
        }
        content = parts
    }

    messages = append(messages, deepSeekRequestMessage{
        Role:    "user",
        Content: content,
    })

    body, err := json.Marshal(deepSeekRequest{
        Model:    c.model,
        Messages: messages,
        Stream:   false,
    })
    if err != nil {
        return "", fmt.Errorf("marshal deepseek request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return "", fmt.Errorf("create deepseek request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.apiKey)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("deepseek request: %w", err)
    }
    defer resp.Body.Close()

    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("read deepseek response: %w", err)
    }
    if resp.StatusCode >= 300 {
        return "", fmt.Errorf("deepseek request failed: %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
    }

    var result deepSeekResponse
    if err := json.Unmarshal(responseBody, &result); err != nil {
        return "", fmt.Errorf("decode deepseek response: %w", err)
    }
    if len(result.Choices) == 0 {
        return "", fmt.Errorf("deepseek returned no choices")
    }

    responseText := strings.TrimSpace(result.Choices[0].Message.Content)
    if responseText == "" {
        return "", fmt.Errorf("deepseek returned an empty response")
    }

    return responseText, nil
}
