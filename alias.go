package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var gitFunctions = []string{
	"function g     { git $args }",
	"function ga    { git add $args }",
	"function gaa   { git add . $args }",
	"function gcm 	{ param([Parameter(Mandatory=$true)] [string]$Message) git commit -m $Message }",
	"function gk    { git checkout $args }",
	"function gkb   { git checkout -b $args }",
	"function gpl   { git pull origin $args }",
	"function gp    { git push origin $args }",
	"function gs    { git status $args }",
	"function gl    { git log $args }",
	"function gcg   { git config --global $args }",
	"function gsp   { git stash push $args }",
	"function gspop { git stash pop . $args }",
	"function gss   { git stash show $args }",
	"function gssu  { git stash show --include-untracked $args }",
}

func setupAliases() {
	profilePath, err := getPowerShellProfilePath()
	if err != nil {
		log.Fatalln("Error getting PowerShell profile path:", err)
	}

	fmt.Println("Setting up aliases in PowerShell profile:", profilePath)
	addAliasesToProfile(profilePath)
}

func addAliasesToProfile(profilePath string) error {
	fileData, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("error reading PowerShell profile: %w", err)
	}
	lines := strings.Split(string(fileData), "\n")

	if strings.Contains(string(fileData), "Git Aliases") {
		fmt.Println("Aliases already exist in PowerShell profile.")
		return nil
	}

	insertIndex := len(lines)
	for i, line := range lines {
		if strings.Contains(line, "oh-my-posh init pwsh") {
			insertIndex = i
			break
		}
	}

	var aliasLines []string
	aliasLines = append(aliasLines, "\n# Git Aliases")
	aliasLines = append(aliasLines, gitFunctions...)

	// insert all existing lines before insertIndex
	newLines := append([]string{}, lines[:insertIndex]...)
	newLines = append(newLines, aliasLines...)
	newLines = append(newLines, "\n")
	// append the final line (should be ohmyposh init)
	newLines = append(newLines, lines[insertIndex:]...)

	if err := os.WriteFile(profilePath, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return fmt.Errorf("error writing to PowerShell profile: %w", err)
	}

	fmt.Println("Aliases added successfully.")
	return nil
}
