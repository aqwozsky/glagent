package main

import (
	"flag"
	"fmt"
	"os"

	glagentgui "glagent/src/modules/glagentGui"
	"glagent/src/modules/installer"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		if err := runSetup(os.Args[2:]); err != nil {
			fmt.Printf("GlAgent setup failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

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

func runSetup(args []string) error {
	setupFlags := flag.NewFlagSet("setup", flag.ContinueOnError)
	systemInstall := setupFlags.Bool("system", false, "install into Program Files and update machine PATH")
	installDir := setupFlags.String("install-dir", "", "override the install directory")
	binaryName := setupFlags.String("binary-name", "glagent.exe", "installed executable name")

	if err := setupFlags.Parse(args); err != nil {
		return err
	}

	scope := installer.ScopeUser
	if *systemInstall {
		scope = installer.ScopeSystem
	}

	result, err := installer.Run(installer.Options{
		Scope:      scope,
		InstallDir: *installDir,
		BinaryName: *binaryName,
	})
	if err != nil {
		return err
	}

	fmt.Println("GlAgent setup completed.")
	fmt.Printf("Scope: %s\n", result.Scope)
	fmt.Printf("Install Directory: %s\n", result.InstallDir)
	fmt.Printf("Installed Binary: %s\n", result.BinaryPath)
	if result.PathUpdated {
		fmt.Println("PATH: updated")
	} else {
		fmt.Println("PATH: already contained install directory")
	}
	fmt.Println("You may need to open a new terminal for PATH changes to take effect.")
	fmt.Printf("Run later with: %s\n", result.BinaryPath)
	return nil
}
