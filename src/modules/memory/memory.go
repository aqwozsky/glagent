package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const memoryFile = "memory.json"

// Item represents a single saved memory entry.
type Item struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Store manages persistent memory storage.
type Store struct {
	Items []Item `json:"items"`
	mu    sync.Mutex
}

var globalStore *Store

// Load reads memory from disk, or creates an empty store.
func Load() *Store {
	if globalStore != nil {
		return globalStore
	}

	store := &Store{}
	data, err := os.ReadFile(memoryFile)
	if err == nil {
		_ = json.Unmarshal(data, store)
	}
	globalStore = store
	return store
}

// save writes the current memory to disk.
func (s *Store) save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(memoryFile, data, 0644)
}

// Add stores a new memory item and persists to disk.
func (s *Store) Add(content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Items = append(s.Items, Item{
		Content:   content,
		CreatedAt: time.Now(),
	})
	return s.save()
}

// Remove deletes a memory item by index (0-based) and persists.
func (s *Store) Remove(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.Items) {
		return fmt.Errorf("invalid index %d (have %d items)", index, len(s.Items))
	}

	s.Items = append(s.Items[:index], s.Items[index+1:]...)
	return s.save()
}

// Clear removes all memory items and persists.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Items = nil
	return s.save()
}

// List returns all stored items.
func (s *Store) List() []Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Items
}

// Count returns the number of stored items.
func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Items)
}

// BuildContext renders all memory items into a string block
// suitable for injecting into a system prompt.
func (s *Store) BuildContext() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Items) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("== User's Saved Memories ==\n")
	b.WriteString("The following facts were explicitly saved by the user. Keep them in mind:\n\n")
	for i, item := range s.Items {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Content))
	}
	b.WriteString("\n== End of Memories ==")
	return b.String()
}

// DetectSaveIntent checks if a user message contains a request to save
// something to memory. Returns the content to save and whether a save
// was detected.
func DetectSaveIntent(msg string) (string, bool) {
	lower := strings.ToLower(msg)

	// Patterns to detect save intent
	prefixes := []string{
		"remember that ",
		"remember this: ",
		"remember: ",
		"save to memory: ",
		"save to memory ",
		"save this to your memory: ",
		"save this to your memory ",
		"save to your memory: ",
		"save to your memory ",
		"memorize: ",
		"memorize this: ",
		"memorize that ",
		"keep in mind: ",
		"keep in mind that ",
		"note: ",
		"note that ",
		"don't forget: ",
		"don't forget that ",
	}

	for _, prefix := range prefixes {
		idx := strings.Index(lower, prefix)
		if idx != -1 {
			content := strings.TrimSpace(msg[idx+len(prefix):])
			if content != "" {
				return content, true
			}
		}
	}

	return "", false
}
