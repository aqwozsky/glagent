package agentMod

import "strings"

type ChatSession struct {
	History  []string
	MaxTurns int
}

func NewChatSession(maxTurns int) *ChatSession {
	return &ChatSession{
		History:  []string{},
		MaxTurns: maxTurns,
	}
}

func (c *ChatSession) AddUserMessage(msg string) {
	c.History = append(c.History, "User: "+msg)
	c.trim()
}

func (c *ChatSession) AddAssistantMessage(msg string) {
	c.History = append(c.History, "Assistant: "+msg)
	c.trim()
}

func (c *ChatSession) BuildPrompt(userInput string) string {
	var b strings.Builder

	if len(c.History) > 0 {
		b.WriteString(strings.Join(c.History, "\n"))
		b.WriteString("\n")
	}

	b.WriteString("User: ")
	b.WriteString(userInput)
	b.WriteString("\nAssistant:")

	return b.String()
}

func (c *ChatSession) Clear() {
	c.History = []string{}
}

func (c *ChatSession) trim() {
	if c.MaxTurns <= 0 {
		return
	}

	maxLines := c.MaxTurns * 2
	if len(c.History) > maxLines {
		c.History = c.History[len(c.History)-maxLines:]
	}
}