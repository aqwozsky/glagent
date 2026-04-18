package main

import (
	"flag"
	"fmt"
	"os"

	glagentgui "glagent/src/modules/glagentGui"
)

func main() {
	continueID := flag.String("continue", "", "resume a previous chat session by id")
	sessionID := flag.String("session", "", "start a new chat session with a custom id")
	flag.Parse()

	options := glagentgui.StartOptions{}
	if *continueID != "" {
		options.ContinueSessionID = *continueID
	} else if *sessionID != "" {
		options.SessionID = *sessionID
	}

	if err := glagentgui.StartGUI(options); err != nil {
		fmt.Printf("GlAgent failed to start: %v\n", err)
		os.Exit(1)
	}
}
