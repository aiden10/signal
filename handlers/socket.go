package handlers

import (
    "encoding/json"
    "log"
    "time"

    "github.com/gorilla/websocket"
    "signal/models"
)

type SocketHandler struct {
    URL          string
    Conn         *websocket.Conn
    EventHandler *EventHandler
    Logger       *log.Logger
    Dialer       *websocket.Dialer
}

func NewSocketHandler(url string, eventHandler *EventHandler, logger *log.Logger) *SocketHandler {
    return &SocketHandler{
        URL:          url,
        EventHandler: eventHandler,
        Logger:       logger,
        Dialer:       websocket.DefaultDialer,
    }
}

func (s *SocketHandler) connect() error {
    conn, _, err := s.Dialer.Dial(s.URL, nil)
    if err != nil {
        return err
    }
    s.Conn = conn
    s.Logger.Printf("connected to websocket: %s", s.URL)
    return nil
}

func (s *SocketHandler) Start() {
    for {
        if s.Conn == nil {
            if err := s.connect(); err != nil {
                s.Logger.Printf("websocket dial failed: %v", err)
                time.Sleep(2 * time.Second)
                continue
            }
        }

        msgType, payload, err := s.Conn.ReadMessage()
        if err != nil {
            if closeErr, ok := err.(*websocket.CloseError); ok {
                s.Logger.Printf("websocket close: code=%d text=%s", closeErr.Code, closeErr.Text)
            } else {
                s.Logger.Printf("websocket read failed: %v", err)
            }
            _ = s.Conn.Close()
            s.Conn = nil
            time.Sleep(1 * time.Second)
            continue
        }

        s.Logger.Printf("raw ws message type=%d bytes=%d payload=%s", msgType, len(payload), string(payload))

        var msg models.SignalEnvelope
        if err := json.Unmarshal(payload, &msg); err != nil {
            s.Logger.Printf("unmarshal failed: %v", err)
            continue
        }

        groupID, author, text, isData, isReceipt := msg.Normalized()
        s.Logger.Printf("normalized: isData=%v isReceipt=%v groupID=%q author=%q text=%q", isData, isReceipt, groupID, author, text)

        switch {
        case isData:
            s.EventHandler.HandleDataMessage(groupID, author, text)
        case isReceipt:
            s.Logger.Printf("receipt event received")
        default:
            s.Logger.Printf("ignored event (not data/receipt)")
        }
    }
}