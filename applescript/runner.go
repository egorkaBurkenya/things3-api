package applescript

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes an AppleScript and returns its output.
func Run(script string) (string, error) {
	f, err := os.CreateTemp("", "things3-*.applescript")
	if err != nil {
		return "", fmt.Errorf("failed to create temp script: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(script); err != nil {
		f.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	f.Close()

	out, err := exec.Command("osascript", f.Name()).CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(out))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("applescript error: %s", errMsg)
	}
	return strings.TrimSpace(string(out)), nil
}

// EscapeString escapes a string for safe embedding in AppleScript.
// Prevents AppleScript injection by escaping backslashes and double quotes.
func EscapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// IsThings3Running checks if Things 3 is currently running.
func IsThings3Running() bool {
	out, err := exec.Command("osascript", "-e",
		`tell application "System Events" to (name of processes) contains "Things3"`).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}
