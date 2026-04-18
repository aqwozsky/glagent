package agentMod

import (
	"strings"
)

type ChatEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatSession struct {
	Entries  []ChatEntry `json:"entries"`
	MaxTurns int         `json:"max_turns"`
}

func NewChatSession(maxTurns int) *ChatSession {
	return &ChatSession{
		Entries:  []ChatEntry{},
		MaxTurns: maxTurns,
	}
}

func NewChatSessionFromEntries(entries []ChatEntry, maxTurns int) *ChatSession {
	session := NewChatSession(maxTurns)
	session.Entries = append(session.Entries, entries...)
	session.trim()
	return session
}

func (c *ChatSession) AddUserMessage(msg string) {
	c.AddMessage("User", msg)
}

func (c *ChatSession) AddAssistantMessage(msg string) {
	c.AddMessage("Assistant", msg)
}

func (c *ChatSession) AddSystemMessage(msg string) {
	c.AddMessage("System", msg)
}

func (c *ChatSession) AddMessage(role, msg string) {
	c.Entries = append(c.Entries, ChatEntry{
		Role:    normalizeRole(role),
		Content: msg,
	})
	c.trim()
}

func (c *ChatSession) BuildPrompt(userInput string) string {
	var b strings.Builder

	for _, entry := range c.Entries {
		b.WriteString(normalizeRole(entry.Role))
		b.WriteString(": ")
		b.WriteString(entry.Content)
		b.WriteString("\n")
	}

	b.WriteString("User: ")
	b.WriteString(userInput)
	b.WriteString("\nAssistant:")

	return b.String()
}

func (c *ChatSession) Clear() {
	c.Entries = []ChatEntry{}
}

func (c *ChatSession) HistoryCount() int {
	return len(c.Entries)
}

func (c *ChatSession) trim() {
	if c.MaxTurns <= 0 {
		return
	}

	maxEntries := c.MaxTurns * 4
	if len(c.Entries) > maxEntries {
		c.Entries = c.Entries[len(c.Entries)-maxEntries:]
	}
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "assistant":
		return "Assistant"
	case "system":
		return "System"
	default:
		return "User"
	}
}
