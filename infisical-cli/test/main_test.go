package tests

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup
	fmt.Println("Setting up CLI...")
	SetupCli()
	// A previous run's `init` leaves .infisical.yaml behind, which changes
	// the init prompts and deadlocks UserInitCmd below. The suite owns this
	// file (init creates it in the working directory), so start clean.
	// (User-level CLI config in $HOME is deliberately left alone.)
	_ = os.Remove(".infisical.yaml")
	fmt.Println("Performing user login...")
	UserLoginCmd()
	fmt.Println("Performing infisical init...")
	UserInitCmd()

	// Run the tests
	code := m.Run()

	// Exit
	os.Exit(code)
}
