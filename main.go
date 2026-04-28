package main

import (
    "log"
    "os"
    "strings"
    "context"

    "signal/handlers"
    "signal/models"
)

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
    
    ctx := context.Background()
    llm, err := NewLLMProvider(ctx, os.Getenv("GEMINI_API_KEY"), os.Getenv("LLM_MODEL"), os.Getenv("EMBEDDING_MODEL"))
    if err != nil {
        logger.Fatal(err)
    }

    sender := models.NewSignalClient(serverURL, phone, targetGroup, logger)

    eventHandler := handlers.NewEventHandler(history, llm, sender, targetGroup, phone)
    socketHandler := handlers.NewSocketHandler(socketURL, eventHandler, logger)

    logger.Println("starting socket handler...")
    socketHandler.Start()
}