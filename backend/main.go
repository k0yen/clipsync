package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

// Data Structures
type IncomingPayload struct {
	Content string `json:"content"`
	IV      string `json:"iv"`
}

type OutgoingSnippet struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	IV        string `json:"iv"`
	Timestamp string `json:"timestamp"`
}

// Room Manager
type Room struct {
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
}

type Hub struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

var (
	hub      *Hub
	db       *sql.DB
	upgrader = websocket.Upgrader{
		ReadBufferSize:	1024,
		WriteBufferSize:	1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
)

func init() {
	hub = &Hub{
		rooms: make(map[string]*Room),
	}
}

// Database Functions
func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./clipsync.db")
	if err != nil {
		return err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS snippets (
		room_id TEXT NOT NULL,
		snippet_id TEXT PRIMARY KEY,
		ciphertext TEXT NOT NULL,
		iv TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_room_created ON snippets(room_id, created_at DESC);
	`
	_, err = db.Exec(schema)
	return err
}

func saveSnippet(roomID, snippetID, content, iv string, timestamp time.Time) error {
	_, err := db.Exec(
		"INSERT INTO snippets (room_id, snippet_id, ciphertext, iv, created_at) VALUES (?, ?, ?, ?, ?)",
		roomID, snippetID, content, iv, timestamp,
	)
	return err
}

func getRecentSnippets(roomID string, limit int) ([]OutgoingSnippet, error) {
	rows, err := db.Query(
		"SELECT snippet_id, ciphertext, iv, created_at FROM snippets WHERE room_id = ? ORDER BY created_at DESC LIMIT ?",
		roomID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []OutgoingSnippet
	for rows.Next() {
		var s OutgoingSnippet
		var ts time.Time
		err := rows.Scan(&s.ID, &s.Content, &s.IV, &ts)
		if err != nil {
			return nil, err
		}
		s.Timestamp = ts.Format(time.RFC3339)
		snippets = append(snippets, s)
	}

	// Reverse to send oldest first
	for i, j := 0, len(snippets)-1; i < j; i, j = i+1, j-1 {
		snippets[i], snippets[j] = snippets[j], snippets[i]
	}

	return snippets, nil
}

func cleanupOldSnippets() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-24 * time.Hour)
		result, err := db.Exec("DELETE FROM snippets WHERE created_at < ?", cutoff)
		if err != nil {
			log.Printf("Cleanup error: %v", err)
			continue
		}
		rows, _ := result.RowsAffected()
		if rows > 0 {
			log.Printf("Cleaned up %d old snippets", rows)
		}
	}
}

// Room Management
func (h *Hub) getOrCreateRoom(roomID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if !exists {
		room = &Room{
			clients: make(map[*websocket.Conn]bool),
		}
		h.rooms[roomID] = room
	}
	return room
}

func (r *Room) addClient(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[conn] = true
}

func (r *Room) removeClient(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, conn)
	conn.Close()
}

func (r *Room) broadcast(snippet OutgoingSnippet) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.Marshal(snippet)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return
	}

	// Broadcast concurrently to avoid blocking
	var wg sync.WaitGroup
	for client := range r.clients {
		wg.Add(1)
		go func(c *websocket.Conn) {
			defer wg.Done()
			if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Write error: %v", err)
			}
		}(client)
	}
	wg.Wait()
}

// WebSocket Handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomID"]

	if roomID == "" {
		http.Error(w, "Room ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	room := hub.getOrCreateRoom(roomID)
	room.addClient(conn)
	defer room.removeClient(conn)

	log.Printf("Client connected to room: %s", roomID)

	// STEP 2: Send history on connect
	history, err := getRecentSnippets(roomID, 50)
	if err != nil {
		log.Printf("History fetch error: %v", err)
	} else if len(history) > 0 {
		historyData, _ := json.Marshal(history)
		conn.WriteMessage(websocket.TextMessage, historyData)
	}

	// Message loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// STEP 3: Parse incoming payload
		var payload IncomingPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			log.Printf("Invalid JSON: %v", err)
			continue
		}

		// Basic validation
		if payload.Content == "" || payload.IV == "" {
			log.Printf("Missing content or IV")
			continue
		}

		// Length check (10MB limit for base64 ciphertext)
		if len(payload.Content) > 10*1024*1024 {
			log.Printf("Content too large")
			continue
		}

		// Generate server-side metadata
		snippetID := uuid.New().String()
		timestamp := time.Now()

		// STEP 3: Persist to database
		if err := saveSnippet(roomID, snippetID, payload.Content, payload.IV, timestamp); err != nil {
			log.Printf("Save error: %v", err)
			continue
		}

		// STEP 4: Broadcast to all clients in room
		outgoing := OutgoingSnippet{
			ID:        snippetID,
			Content:   payload.Content,
			IV:        payload.IV,
			Timestamp: timestamp.Format(time.RFC3339),
		}

		room.broadcast(outgoing)
	}
}

func main() {
	// Initialize database
	if err := initDB(); err != nil {
		log.Fatalf("Database init failed: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized")

	// Start cleanup goroutine
	go cleanupOldSnippets()

	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/ws/{roomID}", handleWebSocket)

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	port := ":8080"
	log.Printf("ClipSync E2EE Backend starting on %s", port)
	log.Fatal(http.ListenAndServe(port, r))
}

