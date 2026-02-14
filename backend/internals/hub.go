package internals

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages all rooms and connections
type Hub struct {
	rooms        map[string]*Room
	mu           sync.RWMutex
	totalClients int
	maxClients   int
}

// Room represents a collection of connected clients
type Room struct {
	ID      string
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
}

// NewHub creates a new hub with a maximum connection limit
func NewHub(maxConnections int) *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		maxClients: maxConnections,
	}
}

// GetOrCreateRoom returns an existing room or creates a new one
func (h *Hub) GetOrCreateRoom(roomID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if !exists {
		room = &Room{
			ID:      roomID,
			clients: make(map[*websocket.Conn]bool),
		}
		h.rooms[roomID] = room
		slog.Info("room created", "room_id", roomID)
	}
	return room
}

// AddClient adds a client to a room
func (h *Hub) AddClient(roomID string, conn *websocket.Conn) error {
	h.mu.Lock()
	if h.totalClients >= h.maxClients {
		h.mu.Unlock()
		return ErrMaxConnections
	}
	h.totalClients++
	h.mu.Unlock()

	room := h.GetOrCreateRoom(roomID)
	room.addClient(conn)

	slog.Info("client connected", "room_id", roomID, "total_clients", h.totalClients)
	return nil
}

// RemoveClient removes a client from a room
func (h *Hub) RemoveClient(roomID string, conn *websocket.Conn) {
	h.mu.Lock()
	h.totalClients--
	totalClients := h.totalClients
	h.mu.Unlock()

	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if exists {
		room.removeClient(conn)

		// Clean up empty rooms
		if room.ClientCount() == 0 {
			h.mu.Lock()
			delete(h.rooms, roomID)
			h.mu.Unlock()
			slog.Info("empty room removed", "room_id", roomID)
		}
	}

	slog.Info("client disconnected", "room_id", roomID, "total_clients", totalClients)
}

// Stats returns hub statistics
func (h *Hub) Stats() (roomCount, clientCount int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms), h.totalClients
}

// addClient adds a websocket connection to the room
func (r *Room) addClient(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[conn] = true
}

// removeClient removes a websocket connection from the room
func (r *Room) removeClient(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, conn)
	conn.Close()
}

// Broadcast sends a snippet to all clients in the room concurrently
func (r *Room) Broadcast(snippet OutgoingSnippet) error {
	data, err := json.Marshal(snippet)
	if err != nil {
		return err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Broadcast concurrently to avoid blocking
	var wg sync.WaitGroup
	for client := range r.clients {
		wg.Add(1)
		go func(c *websocket.Conn) {
			defer wg.Done()
			if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
				slog.Warn("failed to write to client", "error", err)
			}
		}(client)
	}
	wg.Wait()

	return nil
}

// SendHistory sends historical snippets to a specific client
func (r *Room) SendHistory(conn *websocket.Conn, snippets []Snippet) error {
	if len(snippets) == 0 {
		return nil
	}

	// Convert to outgoing format
	outgoing := make([]OutgoingSnippet, len(snippets))
	for i, s := range snippets {
		outgoing[i] = s.ToOutgoing()
	}

	data, err := json.Marshal(outgoing)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// ClientCount returns the number of connected clients
func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

var (
	ErrMaxConnections = &AppError{
		Code:    "MAX_CONNECTIONS",
		Message: "maximum connections reached",
	}
)

// AppError represents an application-specific error
type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}
