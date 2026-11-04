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

type Config struct {
    GeminiAPIKey    string
    DeepSeekAPIKey  string
    DeepSeekModel   string
    DeepSeekBaseURL string
    EmbeddingModel  string
}

type LLMProvider struct {
    client         *genai.Client
    deepSeek       *DeepSeekClient
    EmbeddingModel string
}

func NewLLMProvider(ctx context.Context, config Config) (*LLMProvider, error) {
    if strings.TrimSpace(config.DeepSeekAPIKey) == "" {
        return nil, fmt.Errorf("DEEPSEEK_API_KEY is not set")
    }

    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  config.GeminiAPIKey,
        Backend: genai.BackendGeminiAPI,
    })
    if err != nil {
        return nil, err
    }

    return &LLMProvider{
        client:         client,
        deepSeek:       NewDeepSeekClient(config.DeepSeekAPIKey, config.DeepSeekModel, config.DeepSeekBaseURL),
        EmbeddingModel: config.EmbeddingModel,
    }, nil
}

func (p *LLMProvider) GenerateResponse(relevantMessages, recentMessages []models.Message, systemPrompt, initialPrompt string) string {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    prompt := buildPrompt(relevantMessages, recentMessages, initialPrompt)
    response, err := p.deepSeek.GenerateResponse(ctx, systemPrompt, prompt)
    if err != nil {
        return fmt.Sprintf("Error generating response: %v", err)
    }

    log.Println("Generated Response: " + response)
    return response
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