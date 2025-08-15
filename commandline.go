package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func setupCommandLine() {
	err := installOhMyPosh()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("OhMyPosh successfully installed")

	configPath, err := downloadOhMyPoshConfig()
	if err != nil {
		log.Fatalln(err)
	}

	err = updateTerminalSettings()
	if err != nil {
		log.Fatalln(err)
	}

	err = setupPowerShellProfile(configPath)
	if err != nil {
		log.Fatalln(err)
	}
}

func installOhMyPosh() error {
	commandString := "winget install JanDeDobbeleer.OhMyPosh --source winget --scope user --force"
	cmd := exec.Command("powershell", "-Command", commandString)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func downloadOhMyPoshConfig() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home dir: %w", err)
	}
	configPath := filepath.Join(homeDir, "slimfat.omp.json")
	url := "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/slimfat.omp.json"

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error downloading oh-my-posh config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download config, status: %s", resp.Status)
	}

	out, err := os.Create(configPath)
	if err != nil {
		return "", fmt.Errorf("error creating config file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error saving config file: %w", err)
	}

	fmt.Println("Downloaded Oh My Posh config to:", configPath)
	return configPath, nil
}

func updateTerminalSettings() error {
	settings, err := readTerminalSettings()
	if err != nil {
		return err
	}

	profiles, ok := settings["profiles"].(map[string]any)
	if !ok {
		profiles = map[string]any{}
		settings["profiles"] = profiles
	}

	defaults, ok := profiles["defaults"].(map[string]any)
	if !ok {
		defaults = map[string]any{}
		profiles["defaults"] = defaults
	}

	font, ok := defaults["font"].(map[string]any)
	if !ok {
		font = map[string]any{}
		defaults["font"] = font
	}

	font["face"] = "JetBrainsMono Nerd Font"

	err = writeTerminalSettings(settings)
	if err != nil {
		return err
	}

	fmt.Println("Terminal settings updated")
	return nil
}

func readTerminalSettings() (map[string]any, error) {
	settingsPath := filepath.Join(
		os.Getenv("LOCALAPPDATA"),
		"Packages",
		"Microsoft.WindowsTerminal_8wekyb3d8bbwe",
		"LocalState",
		"settings.json",
	)

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading terminal settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("error unmarshaling terminal settings: %w", err)
	}

	return settings, nil
}

func writeTerminalSettings(settings map[string]any) error {
	settingsPath := filepath.Join(
		os.Getenv("LOCALAPPDATA"),
		"Packages",
		"Microsoft.WindowsTerminal_8wekyb3d8bbwe",
		"LocalState",
		"settings.json",
	)

	newData, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshaling terminal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, newData, 0644); err != nil {
		return fmt.Errorf("error writing terminal settings: %w", err)
	}

	return nil
}

func setupPowerShellProfile(configPath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home dir: %w", err)
	}

	profilePath := filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")

	// check if ps profile exists
	_, err = os.Stat(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("PowerShell profile doesn't exist yet, it will now be created.")
		} else {
			return fmt.Errorf("error checking for PowerShell profile: %w", err)
		}
	} else {
		fmt.Println("PowerShell profile already exists!")
	}

	ohMyPoshLine := fmt.Sprintf("oh-my-posh init pwsh --eval --config %s | Invoke-Expression", configPath)

	var lines []string
	if data, err := os.ReadFile(profilePath); err == nil {
		for line := range strings.SplitSeq(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "oh-my-posh init pwsh") {
				lines = append(lines, line)
			}
		}
	}

	lines = append(lines, ohMyPoshLine)

	err = os.WriteFile(profilePath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("error writing to PowerShell profile: %w", err)
	}

	fmt.Println("PowerShell Profile successfully setup.")
	return nil
}
