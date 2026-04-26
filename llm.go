package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "signal/handlers"
)

type LLMProvider struct {
    Key      string
    Endpoint string
}

func NewLLMProvider(key, endpoint string) *LLMProvider {
    return &LLMProvider{
        Key:      key,
        Endpoint: endpoint,
    }
}

func (provider *LLMProvider) GenerateResponse(context []handlers.Message, initialPrompt string) string {
    if provider.Endpoint == "" {
        return "LLM endpoint is not configured."
    }

    prompt := buildPrompt(context, initialPrompt)

    reqBody := map[string]any{
        "contents": []map[string]any{
            {
                "parts": []map[string]string{
                    {"text": prompt},
                },
            },
        },
    }

    raw, err := json.Marshal(reqBody)
    if err != nil {
        return "I couldn't serialize the LLM request."
    }

    url := provider.Endpoint
    if provider.Key != "" && !strings.Contains(url, "key=") {
        sep := "?"
        if strings.Contains(url, "?") {
            sep = "&"
        }
        url += sep + "key=" + provider.Key
    }

    req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
    if err != nil {
        return "I couldn't create the LLM request."
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 25 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "I couldn't reach the LLM service."
    }
    defer resp.Body.Close()

    var out struct {
        Candidates []struct {
            Content struct {
                Parts []struct {
                    Text string `json:"text"`
                } `json:"parts"`
            } `json:"content"`
        } `json:"candidates"`
        Error *struct {
            Message string `json:"message"`
        } `json:"error"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return "I couldn't parse the LLM response."
    }
    if out.Error != nil && out.Error.Message != "" {
        return fmt.Sprintf("LLM error: %s", out.Error.Message)
    }
    if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
        return "No response came back from the LLM."
    }

    return strings.TrimSpace(out.Candidates[0].Content.Parts[0].Text)
}

func buildPrompt(context []handlers.Message, initialPrompt string) string {
    var b strings.Builder
    b.WriteString("You are responding in a group chat. Keep responses under 2000 characters.\n\n")
    b.WriteString("Recent messages:\n")
    for _, m := range context {
        b.WriteString("- ")
        b.WriteString(m.Author)
        b.WriteString(": ")
        b.WriteString(m.Text)
        b.WriteString("\n")
    }
    b.WriteString("\nUser request:\n")
    b.WriteString(initialPrompt)
    return b.String()
}