package internals

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(hub *Hub, store *Store, cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomID := vars["roomID"]

		if roomID == "" {
			http.Error(w, "room ID required", http.StatusBadRequest)
			return
		}

		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("websocket upgrade failed", "error", err)
			return
		}

		// Add client to room
		if err := hub.AddClient(roomID, conn); err != nil {
			slog.Error("failed to add client", "error", err)
			conn.Close()
			return
		}
		defer hub.RemoveClient(roomID, conn)

		// Send history
		sendHistory(conn, store, roomID, cfg.HistoryLimit)

		// Process messages
		processMessages(conn, hub, store, roomID, cfg.MaxContentSize)
	}
}

func sendHistory(conn *websocket.Conn, store *Store, roomID string, limit int) {
	snippets, err := store.GetRecentSnippets(roomID, limit)
	if err != nil {
		slog.Error("failed to fetch history", "room_id", roomID, "error", err)
		return
	}

	if len(snippets) == 0 {
		return
	}

	// Convert to outgoing format
	outgoing := make([]OutgoingSnippet, len(snippets))
	for i, s := range snippets {
		outgoing[i] = s.ToOutgoing()
	}

	data, err := json.Marshal(outgoing)
	if err != nil {
		slog.Error("failed to marshal history", "error", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		slog.Error("failed to send history", "error", err)
	} else {
		slog.Info("history sent", "room_id", roomID, "count", len(snippets))
	}
}

func processMessages(conn *websocket.Conn, hub *Hub, store *Store, roomID string, maxSize int64) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("websocket error", "room_id", roomID, "error", err)
			}
			break
		}

		// Parse and validate payload
		var payload IncomingPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			slog.Warn("invalid JSON", "room_id", roomID, "error", err)
			continue
		}

		if err := payload.Validate(maxSize); err != nil {
			slog.Warn("validation failed", "room_id", roomID, "error", err)
			continue
		}

		// Create snippet
		snippet := &Snippet{
			RoomID:     roomID,
			SnippetID:  uuid.New().String(),
			Ciphertext: payload.Content,
			IV:         payload.IV,
			CreatedAt:  time.Now(),
		}

		// Save to database
		if err := store.SaveSnippet(snippet); err != nil {
			slog.Error("failed to save snippet", "room_id", roomID, "error", err)
			continue
		}

		// Broadcast to room
		room := hub.GetOrCreateRoom(roomID)
		if err := room.Broadcast(snippet.ToOutgoing()); err != nil {
			slog.Error("broadcast failed", "room_id", roomID, "error", err)
		}
	}
}

// HealthHandler handles health check requests
func HealthHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(); err != nil {
			slog.Error("health check failed", "error", err)
			http.Error(w, "database unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// StatsHandler returns server statistics
func StatsHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomCount, clientCount := hub.Stats()

		stats := map[string]int{
			"rooms":   roomCount,
			"clients": clientCount,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
