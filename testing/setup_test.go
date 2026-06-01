package testing

import (
    "context"
    "log"
    "os"
    "os/exec"
    "testing"
    "time"
    "net"
    "fmt"
    "strings"

    "github.com/joho/godotenv"
    "signal/handlers"
    "signal/models"
    "signal/llm"
    "signal/db"
)

var sshTunnel *exec.Cmd

func waitForTunnel(addr string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        conn, err := net.DialTimeout("tcp", addr, time.Second)
        if err == nil {
            conn.Close()
            return nil
        }
        time.Sleep(500 * time.Millisecond)
    }
    return fmt.Errorf("tunnel at %s did not become ready in time", addr)
}

func TestMain(m *testing.M) {
    if err := godotenv.Load("../.env"); err != nil {
        log.Fatal("Error loading .env: ", err)
    }

    sshCommand := os.Getenv("TESTING_SSH_COMMAND")
    if sshCommand == "" {
        log.Fatal("TESTING_SSH_COMMAND not set in .env")
    }

    parts := strings.Fields(sshCommand)
    sshTunnel = exec.Command(parts[0], parts[1:]...)
    
    if err := sshTunnel.Start(); err != nil {
        log.Fatal("Failed to start SSH tunnel: ", err)
    }
    log.Println("Waiting for SSH tunnel...")

    if err := waitForTunnel("localhost:8080", 10*time.Second); err != nil {
        sshTunnel.Process.Kill()
        log.Fatal("SSH tunnel never became ready: ", err)
    }
    log.Println("SSH tunnel established")

    code := m.Run()
    sshTunnel.Process.Kill()
    os.Exit(code)
}

type TestSuite struct {
    EventHandler *handlers.EventHandler
    Sender       *models.SignalClient
    History      *handlers.HistoryHandler
    LLM          *llm.LLMProvider
    Ctx          context.Context
    TargetGroup  string
    Phone        string
}

func newTestSuite(t *testing.T) *TestSuite {
    t.Helper()
    ctx := context.Background()
    logger := log.Default()
    serverURL := os.Getenv("TESTING_SERVER_URL")
    phone := os.Getenv("TESTING_PHONE_NUMBER")
    targetGroup := Data.TestGroupId

    sender := models.NewSignalClient(serverURL, phone, targetGroup, logger)
    database, err := db.NewDatabase("../test_memory.db")
    if err != nil {
        log.Fatal("Failed to open database: ", err)
    }
    if err := database.Migrate(); err != nil {
        log.Fatal("Failed to migrate database: ", err)
    }    
    
    history := handlers.NewHistoryHandler(database)
    llmProvider, err := llm.NewLLMProvider(ctx, os.Getenv("GEMINI_API_KEY"), os.Getenv("LLM_MODEL"), os.Getenv("EMBEDDING_MODEL"))
    
    if err != nil {
        t.Fatalf("failed to create LLM provider: %v", err)
    }
    
    return &TestSuite{
        EventHandler: handlers.NewEventHandler(history, llmProvider, sender, targetGroup, phone),
        Sender:       sender,
        History:      history,
        LLM:          llmProvider,
        Ctx:          ctx,
        TargetGroup:  targetGroup,
        Phone:        phone,
    }
}