package db

import (
    "database/sql"
	"encoding/binary"
    "math"
	"time"
	"log"

	"signal/models"

	_ "modernc.org/sqlite"
)

type Database struct {
    conn *sql.DB
}

type Memory struct {
    ID      int
	Author string
    Content string
    Vector  []float32
	CreatedAt time.Time
}

func NewDatabase(path string) (*Database, error) {
    conn, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, err
    }
    if err := conn.Ping(); err != nil {
        return nil, err
    }
    return &Database{conn: conn}, nil
}

func (db *Database) Migrate() error {
    _, err := db.conn.Exec(`
        CREATE TABLE IF NOT EXISTS memory (
            id INTEGER PRIMARY KEY,
            group_id TEXT,
			author TEXT,
            content TEXT,
            vector BLOB,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    return err
}

func (db *Database) InsertMemory(groupId, author, content string, vector []float32) error {
    blob := float32SliceToBytes(vector)
    _, err := db.conn.Exec(
        "INSERT INTO memory (group_id, author, content, vector, created_at) VALUES (?, ?, ?, ?, ?)",
        groupId, author, content, blob, time.Now(),
    )
	if err != nil {
        log.Printf("Error creating record: %v", err)
    }
    return err
}

func float32SliceToBytes(v []float32) []byte {
    buf := make([]byte, len(v)*4)
    for i, f := range v {
        binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
    }
    return buf
}

func bytesToFloat32Slice(b []byte) []float32 {
    v := make([]float32, len(b)/4)
    for i := range v {
        v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
    }
    return v
}

func (db *Database) GetAllMemories(groupId string) ([]Memory, error) {
    rows, err := db.conn.Query(
        "SELECT id, author, content, vector, created_at FROM memory WHERE group_id = ?", groupId,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var memories []Memory
    for rows.Next() {
        var m Memory
        var blob []byte
        if err := rows.Scan(&m.ID, &m.Author, &m.Content, &blob, &m.CreatedAt); err != nil {
            return nil, err
        }
        m.Vector = bytesToFloat32Slice(blob)
        memories = append(memories, m)
    }
    return memories, nil
}

func (db *Database) GetRecentMessages(groupId string, n int) []models.Message {
    rows, err := db.conn.Query(
        "SELECT author, content FROM memory WHERE group_id = ? ORDER BY created_at DESC LIMIT ?",
        groupId, n,
    )
    if err != nil {
        log.Printf("Error recent messages: %v", err)
		return []models.Message{}
    }
    defer rows.Close()

    var messages []models.Message
    for rows.Next() {
        var m models.Message
        if err := rows.Scan(&m.Author, &m.Text); err != nil {
            log.Printf("Error recent messages: %v", err)
			return []models.Message{}
        }
        
        messages = append([]models.Message{m}, messages...)
    }
    return messages
}
