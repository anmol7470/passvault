package main

import (
	"fmt"
	"os"

	"github.com/anmol7470/passvault/cmd"
	"github.com/anmol7470/passvault/internal"
)

func main() {
	if err := internal.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := internal.CloseDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close database: %v\n", err)
		}
	}()

	cmd.Execute()
}
