package main

import (
	"fmt"
	"os"
	"runtime"
)

var availableArgs = []string{
	"vscode		| Sets up Visual Studio Code",
	"commandline	| Sets up the command line",
	"alias		| Sets up aliases",
}

func main() {
	if runtime.GOOS != "windows" {
		fmt.Println("This script is only supported on Windows currently.")
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		fail("Missing required argument")
	}
	if len(os.Args) < 2 {
		fail("Only one argument supported")
	}

	arg := os.Args[1]

	switch arg {
	case "vscode":
		setupVscode()
	case "commandline":
		setupCommandLine()
	case "alias":
		setupAliases()
	default:
		fail("Unsupported argument")
	}

	fmt.Println()
	os.Exit(0)
}

func fail(message string) {
	fmt.Printf("%s. The available arguments are: \n\n", message)
	for _, a := range availableArgs {
		fmt.Printf("   - %s\n", a)
	}
	fmt.Println()
	os.Exit(1)
}
