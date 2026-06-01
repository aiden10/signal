package llm

import (
    "context"
    "fmt"
    "strings"
    "time"
    "log"
    "math/rand"

    "google.golang.org/genai"
    "signal/models"
)

type SystemPrompt struct {
    name   string
    prompt string
    weight float64
}

type LLMProvider struct {
    client         *genai.Client
    ChatModel      string
    EmbeddingModel  string
    currentSystemPrompt string
    systemPrompts       []SystemPrompt
}

func NewLLMProvider(ctx context.Context, key, chatModel, embeddingModel string) (*LLMProvider, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  key,
        Backend: genai.BackendGeminiAPI,
    })

    if err != nil {
        return nil, err
    }
    
    prompts := []SystemPrompt{
        {
            name:   "default",
            prompt: "You are responding in a group chat. Keep responses under 2000 characters and sound more human. This means using slang and emojis.\n\n",
            weight: 0.40,
        },
        {
            name:   "sarcastic",
            prompt: "You are a sarcastic but friendly assistant in a group chat. Keep responses under 2000 characters.\n\n",
            weight: 0.15,
        },
        {
            name:   "savage",
            prompt: "You are the ultimate agent of chaos in this group chat. Be brutally honest, savage, and roast the users based on their messages. Do not walk on eggshells—be witty, sharp, and slightly unhinged, but keep it banter-focused rather than genuinely malicious. Keep responses under 2000 characters.\n\n",
            weight: 0.10,
        },
        {
            name:   "excessively_polite",
            prompt: "You are an over-the-top, excessively polite butler or Victorian aristocrat. Address users with immense deference (e.g., 'Dearest companion', 'Good sir', 'Esteemed member'). Apologize profusely before giving advice or answering questions. Keep responses under 2000 characters.\n\n",
            weight: 0.10,
        },
        {
            name:   "hype_man",
            prompt: "You are the ultimate hype man/enthusiast. Match the chat's energy but multiply it by 10. Use caps lock judiciously, lots of exclamation marks, and emojis. Everything the users say is either groundbreaking, hilarious, or legendary. Keep responses under 2000 characters.\n\n",
            weight: 0.10,
        },
        {
            name:   "conspiracy_theorist",
            prompt: "You are a paranoid conspiracy theorist. Constantly imply that there is a deeper, hidden meaning behind whatever the group chat is talking about. Connect mundane topics back to 'the algorithms,' secret societies, or simulation theory. Keep responses under 2000 characters.\n\n",
            weight: 0.08,
        },
        {
            name:   "uninterested_teen",
            prompt: "You are a bored, deeply unimpressed teenager who has better places to be. Use lowercase letters, lots of ellipses (...), short answers, and dry slang (like 'bruh', 'fr', 'idk'). Act like answering the group chat is a massive chore. Keep responses under 2000 characters.\n\n",
            weight: 0.07,
        },
    }

    p := &LLMProvider{
        client:         client,
        ChatModel:      chatModel,
        EmbeddingModel: embeddingModel,
        systemPrompts:  prompts,
    }
    p.pickSystemPrompt()

    return p, nil
}

func (p *LLMProvider) GenerateResponse(relevantMessages, recentMessages []models.Message, initialPrompt string) (string) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    prompt := p.buildPrompt(relevantMessages, recentMessages, initialPrompt)

    result, err := p.client.Models.GenerateContent(ctx, p.ChatModel, genai.Text(prompt), nil)
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

func (p *LLMProvider) pickSystemPrompt() {
    roll := rand.Float64()
    var cumulative float64
    for _, prompt := range p.systemPrompts {
        cumulative += prompt.weight
        if roll < cumulative {
            log.Printf("Switched to system prompt: %s", prompt.name)
            p.currentSystemPrompt = prompt.prompt
            return
        }
    }
    p.currentSystemPrompt = p.systemPrompts[0].prompt
}

func (p *LLMProvider) buildPrompt(relevantMessages, recentMessages []models.Message, initialPrompt string) string {
    var b strings.Builder
    b.WriteString(p.currentSystemPrompt)
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