package memory

import (
	"strings"
	"testing"
	"time"
)

func TestDetectSaveIntentRemember(t *testing.T) {
	content, ok := DetectSaveIntent("remember my name is Baris")
	if !ok {
		t.Fatalf("expected remember intent to be detected")
	}
	if content != "my name is Baris" {
		t.Fatalf("unexpected remembered content: %q", content)
	}
}

func TestBuildContextInterpretsFirstPersonAsUserFacts(t *testing.T) {
	store := &Store{
		Items: []Item{
			{
				ID:        "mem_test_1",
				Content:   "my name is Baris",
				CreatedAt: time.Now(),
			},
			{
				ID:        "mem_test_2",
				Content:   "I use pnpm",
				CreatedAt: time.Now(),
			},
		},
	}

	context := store.BuildContext()
	if !strings.Contains(context, `Raw user statement: "my name is Baris"`) {
		t.Fatalf("expected raw user statement in context, got: %s", context)
	}
	if !strings.Contains(context, "Interpreted user fact: The user's name is Baris") {
		t.Fatalf("expected normalized name memory in context, got: %s", context)
	}
	if !strings.Contains(context, "Interpreted user fact: The user uses pnpm") {
		t.Fatalf("expected normalized tool memory in context, got: %s", context)
	}
}
