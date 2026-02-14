package internals

import (
	"errors"
	"time"
)

// IncomingPayload represents a message from client to server
type IncomingPayload struct {
	Content string `json:"content"` // Base64 Ciphertext
	IV      string `json:"iv"`      // Base64 Initialization Vector
}

// OutgoingSnippet represents a message from server to client
type OutgoingSnippet struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	IV        string `json:"iv"`
	Timestamp string `json:"timestamp"`
}

// Snippet represents a database record
type Snippet struct {
	RoomID     string
	SnippetID  string
	Ciphertext string
	IV         string
	CreatedAt  time.Time
}

// Validation errors
var (
	ErrMissingContent  = errors.New("content field is required")
	ErrMissingIV       = errors.New("iv field is required")
	ErrContentTooLarge = errors.New("content exceeds maximum size")
)

// Validate checks if the incoming payload is valid
func (p *IncomingPayload) Validate(maxSize int64) error {
	if p.Content == "" {
		return ErrMissingContent
	}
	if p.IV == "" {
		return ErrMissingIV
	}
	if int64(len(p.Content)) > maxSize {
		return ErrContentTooLarge
	}
	return nil
}

// ToOutgoing converts a Snippet to OutgoingSnippet
func (s *Snippet) ToOutgoing() OutgoingSnippet {
	return OutgoingSnippet{
		ID:        s.SnippetID,
		Content:   s.Ciphertext,
		IV:        s.IV,
		Timestamp: s.CreatedAt.Format(time.RFC3339),
	}
}
