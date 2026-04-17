package consolemarkdown

import (
	"fmt"
	"os"
	"strings"
	"sync"

	customLogging "glagent/src/modules/customLogging"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var renderer *glamour.TermRenderer
var mu sync.Mutex
var lastRenderedLines int

var boxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1, 2).
	BorderForeground(lipgloss.Color("63"))

func init() {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 100
	}

	renderer, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-6),
	)
	if err != nil {
		panic(err)
	}
}

func PrintMarkdown(text string) {
	mu.Lock()
	defer mu.Unlock()

	final, err := renderMarkdown(text)
	if err != nil {
		customLogging.Error("Markdown render error: " + err.Error())
		return
	}

	fmt.Println(final)
	lastRenderedLines = countLines(final)
}

func LiveRenderMarkdown(text string) {
	mu.Lock()
	defer mu.Unlock()

	final, err := renderMarkdown(text)
	if err != nil {
		customLogging.Error("Markdown render error: " + err.Error())
		return
	}

	clearPreviousRender(lastRenderedLines)
	fmt.Print(final)

	lastRenderedLines = countLines(final)
}

func FinishLiveRender() {
	mu.Lock()
	defer mu.Unlock()

	fmt.Println()
	lastRenderedLines = 0
}

func renderMarkdown(text string) (string, error) {
	out, err := renderer.Render(text)
	if err != nil {
		return "", err
	}

	final := boxStyle.Render(out)
	return final, nil
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func clearPreviousRender(lines int) {
	if lines <= 0 {
		return
	}

	fmt.Print("\r")

	for i := 0; i < lines; i++ {
		fmt.Print("\033[1A")
	}

	for i := 0; i < lines; i++ {
		fmt.Print("\033[2K")
		if i < lines-1 {
			fmt.Print("\033[1B")
		}
	}

	for i := 0; i < lines-1; i++ {
		fmt.Print("\033[1A")
	}

	fmt.Print("\r")
}