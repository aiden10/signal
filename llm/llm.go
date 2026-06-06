package llm

import (
    "context"
    "fmt"
    "strings"
    "time"
    "log"

    "google.golang.org/genai"
    "signal/models"
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

func (p *LLMProvider) GenerateResponse(relevantMessages, recentMessages []models.Message, initialPrompt string) (string) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    prompt := buildPrompt(relevantMessages, recentMessages, initialPrompt)
    config := &genai.GenerateContentConfig{
        Tools: []*genai.Tool{
            {GoogleSearch: &genai.GoogleSearch{}},
        },
    }
    
    result, err := p.client.Models.GenerateContent(ctx, p.ChatModel, genai.Text(prompt), config)

    if err != nil {
        return fmt.Sprintf("Error generating response: %v", err)
    }

    if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
        return "Error parsing result"
    }

    part := result.Candidates[0].Content.Parts[0]
    if part == nil || part.Text == "" {
        return "Unexpected response format"
    }

    log.Println("Generated Response: " + strings.TrimSpace(part.Text))
    return strings.TrimSpace(part.Text)
}

func (p *LLMProvider) GenerateEmbedding(text string) ([]float32, error) {
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

func buildPrompt(relevantMessages, recentMessages []models.Message, initialPrompt string) string {
    var b strings.Builder
    b.WriteString("You are responding in a group chat. Keep responses under 2000 characters and sound more human. Do not inclue any markdown formatting and occasionally use emojis and slang.\n\n")
    b.WriteString("Recent messages:\n")
    for _, m := range recentMessages {
        b.WriteString("- ")
        b.WriteString(m.Author)
        b.WriteString(": ")
        b.WriteString(m.Text)
        b.WriteString("\n")
    }
    b.WriteString("Relevant messages:\n")
    for _, m := range relevantMessages {
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