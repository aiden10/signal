package main

import (
    "context"
    "fmt"
    "strings"

    "google.golang.org/genai"
    "signal/handlers"
)

type LLMProvider struct {
    client         *genai.Client
    ChatModel      string
    EmbeddingModel  string
}

func NewLLMProvider(ctx context.Context, key, chatModel, embeddingModel string) (*LLMProvider, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  key,
        Backend: genai.BackendGeminiAPI,
    })
    if err != nil {
        return nil, err
    }

    return &LLMProvider{
        client:        client,
        ChatModel:     chatModel,
        EmbeddingModel: embeddingModel,
    }, nil
}

func (p *LLMProvider) GenerateResponse(history []handlers.Message, initialPrompt string) (string) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    prompt := buildPrompt(history, initialPrompt)

    result, err := p.client.Models.GenerateContent(ctx, p.ChatModel, genai.Text(prompt), nil)
    if err != nil {
        return fmt.Sprintf("Error generating response: %v", err)
    }

    if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
        return "Error parsing result"
    }

    text, ok := result.Candidates[0].Content.Parts[0].(*genai.Part)
    if !ok {
        return "Unexpected Response format"
    }

    return strings.TrimSpace(text.Text)
}

func (p *LLMProvider) EmbedText(text string) ([]float32, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    contents := []*genai.Content{
        genai.NewContentFromText(text, genai.RoleUser),
    }

    result, err := p.client.Models.EmbedContent(ctx, p.EmbeddingModel, contents, nil)
    if err != nil {
        return nil, err
    }

    if len(result.Embeddings) == 0 {
        return nil, fmt.Errorf("no embeddings returned")
    }

    return result.Embeddings[0].Values, nil
}

func buildPrompt(context []handlers.Message, initialPrompt string) string {
    var b strings.Builder
    b.WriteString("You are responding in a group chat. Keep responses under 2000 characters and sound more human.\n\n")
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