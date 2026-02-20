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

// FindByNameScript returns an AppleScript snippet that finds an object by name
// using iteration with trimming. Things 3 adds trailing spaces to names, and
// the `whose` clause does not work reliably for areas and projects.
// `class` is "area" or "project", `varName` is the variable to assign.
func FindByNameScript(class, varName, nameValue string) string {
	escaped := EscapeString(nameValue)
	return fmt.Sprintf(`	set %s to missing value
	repeat with _item in %ss
		if (name of _item) starts with "%s" then
			set trimmedName to name of _item
			-- trim trailing spaces
			repeat while trimmedName ends with " "
				set trimmedName to text 1 thru -2 of trimmedName
			end repeat
			if trimmedName is "%s" then
				set %s to _item
				exit repeat
			end if
		end if
	end repeat
	if %s is missing value then error "Cannot find %s named \"%s\""`, varName, class, escaped, escaped, varName, varName, class, escaped)
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
