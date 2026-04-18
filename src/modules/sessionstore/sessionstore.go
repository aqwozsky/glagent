package sessionstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"glagent/src/modules/agentMod"
)

type Message struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type Session struct {
	ID             string               `json:"id"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Messages       []Message            `json:"messages"`
	ChatEntries    []agentMod.ChatEntry `json:"chat_entries"`
	PermissionMode string               `json:"permission_mode"`
}

func New(id string) *Session {
	now := time.Now()
	return &Session{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []Message{},
	}
}

func NewID() string {
	return fmt.Sprintf("chat-%s", time.Now().Format("20060102-150405"))
}

func Load(id string) (*Session, error) {
	path := sessionPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("session %q not found", id)
		}
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	if session.ID == "" {
		session.ID = id
	}

	return &session, nil
}

func Save(session *Session) error {
	if session == nil {
		return errors.New("nil session")
	}

	if err := os.MkdirAll(baseDir(), 0755); err != nil {
		return err
	}

	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sessionPath(session.ID), data, 0644)
}

func baseDir() string {
	return filepath.Join(".glagent", "sessions")
}

func sessionPath(id string) string {
	return filepath.Join(baseDir(), id+".json")
}
