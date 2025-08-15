package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	HWND_BROADCAST = 0xffff
	WM_FONTCHANGE  = 0x001D
)

var (
	user32       = windows.NewLazySystemDLL("user32.dll")
	sendMessageW = user32.NewProc("SendMessageW")
)

var extensionIds = []string{
	"fogio.jetbrains-file-icon-theme",
	"chadbaileyvh.oled-pure-black---vscode",
	"esbenp.prettier-vscode",
}

func main() {
	err := InstallExtensions()
	if err != nil {
		log.Fatalln(err)
	}

	err = ConfigureSettings()
	if err != nil {
		log.Fatalln(err)
	}
}

func InstallExtensions() error {
	for _, id := range extensionIds {
		cmd := exec.Command("code", "--install-extension", id)
		fmt.Println("Installing extension: ", id)

		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to install extension [%s] with error: %w", id, err)
		}
		fmt.Println(string(out))
	}
	return nil
}

func ConfigureSettings() error {
	// get font
	fontErr := getFont()
	if fontErr != nil {
		fmt.Printf("error installing JetBrains font: %v", fontErr)
	}

	settingsPath := filepath.Join(os.Getenv("APPDATA"), "Code", "User", "settings.json")

	settings := map[string]any{}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("error reading settings file from %s: %w", settingsPath, err)
	}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		return fmt.Errorf("error unmarshaling vscode settings: %w", err)
	}

	// font
	if fontErr == nil {
		settings["editor.fontFamily"] = "JetBrainsMono Nerd Font"
		settings["editor.fontSize"] = 13
		settings["editor.fontLigatures"] = true
	}

	// theme
	settings["workbench.colorTheme"] = "Dark+ Pure Black (OLED)"

	// autosave
	settings["files.autoSave"] = "afterDelay"

	newData, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal new settings: %w", err)
	}
	err = os.WriteFile(settingsPath, newData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write new settings: %w", err)
	}

	fmt.Println("VS Code settings updated.")
	return nil
}

func getFont() error {
	fontsDir := "C:\\Windows\\Fonts\\"

	// check for font first
	cmd := exec.Command("cmd", "/C", "dir", "/s", fontsDir, "/b", "/o:gn")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error checking if font already installed...\nGoing to install anyway.")
	} else {
		if strings.Contains(string(out), "JetBrainsMono") {
			fmt.Println("JetBrains font already installed, skipping download!")
			return nil
		}
	}

	// res, err := http.Get("https://download.jetbrains.com/fonts/JetBrainsMono-2.304.zip")
	res, err := http.Get("https://github.com/ryanoasis/nerd-fonts/releases/download/v3.4.0/JetBrainsMono.zip")
	if err != nil {
		return fmt.Errorf("failed to download font: %w", err)
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return fmt.Errorf("failed to read font zip: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, file := range zr.File {
		if !file.FileInfo().IsDir() && filepath.Ext(file.Name) == ".ttf" {
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file in zip: %w", err)
			}

			outPath := filepath.Join(fontsDir, filepath.Base(file.Name))
			outFile, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("failed to create font file: %w", err)
			}

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return fmt.Errorf("failed to write font file: %w", err)
			}
			rc.Close()
			outFile.Close()
			fmt.Println("JetBrains font installed at: ", outPath)

			err = registerFontInRegistry(filepath.Base(file.Name))
			if err != nil {
				return fmt.Errorf("failed to register font: %w", err)
			}
		}
	}

	broadcastFontChange()

	fmt.Println("JetBrains fonts installed!")
	return nil
}

func registerFontInRegistry(fontFile string) error {
	displayName := strings.TrimSuffix(fontFile, filepath.Ext(fontFile)) + " (TrueType)"

	k, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts`,
		registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	err = k.SetStringValue(displayName, fontFile)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}
	return nil
}

func broadcastFontChange() {
	sendMessageW.Call(uintptr(HWND_BROADCAST), uintptr(WM_FONTCHANGE), 0, 0)
}
