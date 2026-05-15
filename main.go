//go:generate goversioninfo -64

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"spec-collector/internal/agent"
	"spec-collector/internal/models"
)

func main() {
	m, err := agent.Gather()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering spec: %v\n", err)
		pause()
		os.Exit(1)
	}

	output := agent.FormatSpec(m)
	fmt.Print(output)

	if runtime.GOOS == "windows" {
		if path := agent.GenerateHTML(m); path != "" {
			fmt.Printf("📄 Report: %s\n", path)
			exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", path).Start()
		}
		interactive(m, output)
	} else {
		pause()
	}
}

func interactive(m *models.Machine, output string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n[C] Copy  [S] Save to file  [R] HTML report  [E] Edit  [Enter] Exit: ")
		if !scanner.Scan() {
			return
		}
		cmd := strings.ToLower(strings.TrimSpace(scanner.Text()))
		switch cmd {
		case "c":
			if copyToClipboard(output) {
				fmt.Println("✓ Copied to clipboard!")
			} else {
				fmt.Println("✗ Clipboard failed")
			}
		case "s":
			if path := saveToFile(m, output); path != "" {
				fmt.Printf("✓ Saved to %s\n", path)
			} else {
				fmt.Println("✗ Save failed")
			}
		case "r":
			if path := agent.GenerateHTML(m); path != "" {
				fmt.Printf("✓ Report: %s\n", path)
				exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", path).Start()
			} else {
				fmt.Println("✗ Report failed")
			}
		case "e":
			if editInEditor(output) {
				fmt.Println("✓ Opened in editor")
			} else {
				fmt.Println("✗ Editor failed")
			}
		case "":
			return
		}
	}
}

func copyToClipboard(text string) bool {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("clip")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run() == nil
	default:
		for _, prog := range []string{"xclip", "xsel", "wl-copy"} {
			cmd := exec.Command(prog)
			cmd.Stdin = strings.NewReader(text)
			if cmd.Run() == nil {
				return true
			}
		}
		return false
	}
}

func editInEditor(text string) bool {
	tmp := filepath.Join(os.TempDir(),
		fmt.Sprintf("spec-%s.txt", time.Now().Format("20060102-150405")))
	if err := os.WriteFile(tmp, []byte(text), 0644); err != nil {
		return false
	}

	switch runtime.GOOS {
	case "windows":
		return exec.Command("notepad.exe", tmp).Start() == nil
	default:
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}
		return exec.Command(editor, tmp).Start() == nil
	}
}

func saveToFile(m *models.Machine, output string) string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)
	hostname := m.Hostname
	if hostname == "" {
		hostname = "unknown"
	}
	name := fmt.Sprintf("spec-%s-%s.txt", hostname, time.Now().Format("20060102-150405"))
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(output), 0644); err != nil {
		return ""
	}
	return path
}

func pause() {
	if runtime.GOOS == "windows" {
		fmt.Print("\nNhấn Enter để thoát...")
		fmt.Scanln()
	}
}
