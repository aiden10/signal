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

    if s.EventHandler != nil && s.EventHandler.Sender != nil {
        if setter, ok := any(s.EventHandler.Sender).(interface{ SetConn(*websocket.Conn) }); ok {
            setter.SetConn(conn)
            s.Logger.Printf("sender socket injected")
        }
    }

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
            s.Logger.Printf("websocket read failed (type %d): %v", msgType, err)

            // Clear injected sender conn on disconnect
            if s.EventHandler != nil && s.EventHandler.Sender != nil {
                if setter, ok := any(s.EventHandler.Sender).(interface{ SetConn(*websocket.Conn) }); ok {
                    setter.SetConn(nil)
                }
            }

            _ = s.Conn.Close()
            s.Conn = nil
            time.Sleep(1 * time.Second)
            continue
        }

        var msg models.SignalEnvelope
        if err := json.Unmarshal(payload, &msg); err != nil {
            s.Logger.Printf("unmarshal failed: %v", err)
            continue
        }

        groupID, author, text, isData, isReceipt := msg.Normalized()
        switch {
        case isData:
            s.EventHandler.HandleDataMessage(groupID, author, text)
        case isReceipt:
            s.Logger.Printf("receipt event received")
        }
    }
}