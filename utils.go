package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func getPowerShellProfilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home dir: %w", err)
	}

	profilePath := filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")

	// check if ps profile exists
	_, err = os.Stat(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("PowerShell profile doesn't exist yet, it will now be created.")
		} else {
			return "", fmt.Errorf("error checking for PowerShell profile: %w", err)
		}
	} else {
		fmt.Println("PowerShell profile already exists!")
	}

	return profilePath, nil
}
