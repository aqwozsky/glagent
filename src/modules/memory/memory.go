package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"glagent/src/modules/appstate"
)

// Item represents a single saved memory entry.
type Item struct {
	ID        string    `json:"id"`
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
	data, err := os.ReadFile(appstate.MemoryFilePath())
	if err == nil {
		_ = json.Unmarshal(data, store)
		if store.ensureIDsLocked() {
			_ = store.saveUnlocked()
		}
	}
	globalStore = store
	return store
}

// save writes the current memory to disk.
func (s *Store) save() error {
	if _, err := appstate.EnsureBaseDir(); err != nil {
		return err
	}
	s.ensureIDsLocked()
	return s.saveUnlocked()
}

func (s *Store) saveUnlocked() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(appstate.MemoryFilePath(), data, 0644)
}

// Add stores a new memory item and persists to disk.
func (s *Store) Add(content string) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := Item{
		ID:        newID(),
		Content:   content,
		CreatedAt: time.Now(),
	}
	s.Items = append(s.Items, item)
	return item, s.save()
}

// RemoveByID deletes a memory item by id and persists.
func (s *Store) RemoveByID(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("memory id is required")
	}

	for i, item := range s.Items {
		if item.ID == id {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return s.save()
		}
	}

	return fmt.Errorf("memory id %q not found", id)
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

	items := make([]Item, len(s.Items))
	copy(items, s.Items)
	return items
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
	b.WriteString("== Persistent User Memory ==\n")
	b.WriteString("These facts were explicitly saved by the user. Treat them as persistent context unless the user removes them.\n")
	b.WriteString("They describe the user, the user's preferences, the user's projects, or the user's environment.\n")
	b.WriteString("Important: many memories are written in first person from the user's perspective.\n")
	b.WriteString("Interpret first-person wording as referring to the user, not to GlAgent.\n")
	b.WriteString("Example: if a memory says \"my name is Baris\", that means the user's name is Baris.\n")
	b.WriteString("If the user asks about their name, preferences, setup, or other stored facts and the answer appears below, answer from memory directly.\n")
	b.WriteString("Never claim that the user's name, role, or personal details are your own. You are always GlAgent.\n\n")
	for _, item := range s.Items {
		b.WriteString(fmt.Sprintf("- Memory [%s]\n", shortID(item.ID)))
		b.WriteString(fmt.Sprintf("  Raw user statement: %q\n", item.Content))
		if normalized := normalizeUserFact(item.Content); normalized != "" {
			b.WriteString(fmt.Sprintf("  Interpreted user fact: %s\n", normalized))
		}
	}
	b.WriteString("\n== End of User Memory ==")
	return b.String()
}

// DetectSaveIntent checks if a user message contains a request to save
// something to memory. Returns the content to save and whether a save
// was detected.
func DetectSaveIntent(msg string) (string, bool) {
	lower := strings.ToLower(msg)

	prefixes := []string{
		"remember ",
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
			if strings.HasPrefix(strings.ToLower(content), "that ") {
				content = strings.TrimSpace(content[len("that "):])
			}
			if content != "" {
				return content, true
			}
		}
	}

	return "", false
}

func (s *Store) ensureIDsLocked() bool {
	updated := false
	for i := range s.Items {
		if strings.TrimSpace(s.Items[i].ID) == "" {
			s.Items[i].ID = newID()
			updated = true
		}
		if s.Items[i].CreatedAt.IsZero() {
			s.Items[i].CreatedAt = time.Now()
			updated = true
		}
	}
	return updated
}

func newID() string {
	return "mem_" + time.Now().UTC().Format("20060102_150405.000000000")
}

func shortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[len(id)-12:]
}

func normalizeUserFact(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "my "):
		return "The user's " + strings.TrimSpace(trimmed[3:])
	case strings.HasPrefix(lower, "i am "):
		return "The user is " + strings.TrimSpace(trimmed[5:])
	case strings.HasPrefix(lower, "i'm "):
		return "The user is " + strings.TrimSpace(trimmed[4:])
	case strings.HasPrefix(lower, "im "):
		return "The user is " + strings.TrimSpace(trimmed[3:])
	case strings.HasPrefix(lower, "i use "):
		return "The user uses " + strings.TrimSpace(trimmed[6:])
	case strings.HasPrefix(lower, "i prefer "):
		return "The user prefers " + strings.TrimSpace(trimmed[9:])
	case strings.HasPrefix(lower, "i like "):
		return "The user likes " + strings.TrimSpace(trimmed[7:])
	case strings.HasPrefix(lower, "i work on "):
		return "The user works on " + strings.TrimSpace(trimmed[10:])
	default:
		return "The user said: " + trimmed
	}
}
