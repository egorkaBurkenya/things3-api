package database

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/egorkaBurkenya/things3-api/models"
)

// thingsDBPath returns the path to the Things 3 SQLite database.
func thingsDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	container := filepath.Join(home, "Library", "Group Containers", "JLMPQHK86H.com.culturedcode.ThingsMac")
	entries, err := os.ReadDir(container)
	if err != nil {
		return "", fmt.Errorf("cannot read Things container: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "ThingsData-") {
			dbPath := filepath.Join(container, e.Name(), "Things Database.thingsdatabase", "main.sqlite")
			if _, err := os.Stat(dbPath); err == nil {
				return dbPath, nil
			}
		}
	}
	return "", fmt.Errorf("Things 3 database not found")
}

// query runs a sqlite3 query and returns the output.
func query(sql string) (string, error) {
	dbPath, err := thingsDBPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("sqlite3", "-separator", "\t", dbPath, sql)
	out, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(out))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("sqlite3 error: %s", errMsg)
	}
	return strings.TrimSpace(string(out)), nil
}

// openThingsURL opens a things:/// URL via AppleScript.
func openThingsURL(thingsURL string) error {
	script := fmt.Sprintf(`open location "%s"`, thingsURL)
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(out))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("failed to open things URL: %s", errMsg)
	}
	return nil
}

// GetChecklistItems retrieves all checklist items for a task (via SQLite).
func GetChecklistItems(taskID string) ([]models.ChecklistItem, error) {
	if err := models.ValidateThingsID(taskID); err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(
		`SELECT uuid, title, status FROM TMChecklistItem WHERE task='%s' ORDER BY "index" ASC`,
		escapeSQLite(taskID),
	)

	out, err := query(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get checklist items: %w", err)
	}

	return parseChecklistItems(out), nil
}

// CreateTaskWithChecklist creates a task with checklist items via URL scheme,
// then looks up the created task ID from SQLite.
// Returns the task ID.
func CreateTaskWithChecklist(title string, checklistItems []string, notes, project, area, due, when string, tags []string) (string, error) {
	params := url.Values{}
	params.Set("title", title)
	params.Set("checklist-items", strings.Join(checklistItems, "\n"))
	if notes != "" {
		params.Set("notes", notes)
	}
	if project != "" {
		params.Set("list", project)
	}
	if area != "" && project == "" {
		params.Set("list", area)
	}
	if due != "" {
		params.Set("deadline", due)
	}
	if when != "" {
		params.Set("when", when)
	}
	if len(tags) > 0 {
		params.Set("tags", strings.Join(tags, ","))
	}
	params.Set("show-quick-entry", "false")

	// Things URL scheme expects percent-encoding (%20), not form-encoding (+).
	thingsURL := "things:///add?" + strings.ReplaceAll(params.Encode(), "+", "%20")

	if err := openThingsURL(thingsURL); err != nil {
		return "", err
	}

	// Wait for Things to process the URL and write to database.
	// Use a generous initial wait — Things needs time to process the URL.
	time.Sleep(1500 * time.Millisecond)

	var taskID string
	for i := 0; i < 15; i++ {
		// Query the most recently created task with matching title that has checklist items.
		sql := fmt.Sprintf(
			`SELECT t.uuid FROM TMTask t
			 JOIN TMChecklistItem ci ON ci.task = t.uuid
			 WHERE t.title='%s'
			 GROUP BY t.uuid
			 ORDER BY t.creationDate DESC LIMIT 1`,
			escapeSQLite(title),
		)
		out, err := query(sql)
		if err == nil && strings.TrimSpace(out) != "" {
			taskID = strings.TrimSpace(out)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if taskID == "" {
		return "", fmt.Errorf("task created via URL scheme but could not find ID in database")
	}

	return taskID, nil
}

// AddChecklistItem adds a checklist item to an existing task via URL scheme.
// Requires Things URL Scheme auth token.
func AddChecklistItem(taskID string, title string, authToken string) error {
	if err := models.ValidateThingsID(taskID); err != nil {
		return err
	}

	params := url.Values{}
	params.Set("id", taskID)
	params.Set("append-checklist-items", title)
	if authToken != "" {
		params.Set("auth-token", authToken)
	}

	thingsURL := "things:///update?" + strings.ReplaceAll(params.Encode(), "+", "%20")
	return openThingsURL(thingsURL)
}

// AddChecklistItemDirect adds a checklist item directly via SQLite.
// Used when no auth token is available. Items appear in API reads
// but may not immediately appear in Things UI.
func AddChecklistItemDirect(taskID string, req models.CreateChecklistItemRequest) (*models.ChecklistItem, error) {
	if err := models.ValidateThingsID(taskID); err != nil {
		return nil, err
	}

	uuid := generateUUID()
	now := coreDataTimestamp()

	sql := fmt.Sprintf(
		`INSERT INTO TMChecklistItem (uuid, task, title, status, "index", creationDate, userModificationDate, leavesTombstone)
		VALUES ('%s', '%s', '%s', 0, (SELECT COALESCE(MAX("index"), 0) + 1 FROM TMChecklistItem WHERE task='%s'), %f, %f, 1)`,
		escapeSQLite(uuid),
		escapeSQLite(taskID),
		escapeSQLite(req.Title),
		escapeSQLite(taskID),
		now, now,
	)

	if _, err := query(sql); err != nil {
		return nil, fmt.Errorf("failed to add checklist item: %w", err)
	}

	return &models.ChecklistItem{
		ID:        uuid,
		Title:     req.Title,
		Completed: false,
	}, nil
}

// UpdateChecklistItem updates a checklist item via SQLite.
func UpdateChecklistItem(taskID, itemID string, req models.UpdateChecklistItemRequest) (*models.ChecklistItem, error) {
	if err := models.ValidateThingsID(taskID); err != nil {
		return nil, err
	}
	if err := models.ValidateThingsID(itemID); err != nil {
		return nil, err
	}

	var sets []string
	now := coreDataTimestamp()

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title='%s'", escapeSQLite(*req.Title)))
	}
	if req.Completed != nil {
		if *req.Completed {
			sets = append(sets, "status=3")
			sets = append(sets, fmt.Sprintf("stopDate=%f", now))
		} else {
			sets = append(sets, "status=0")
			sets = append(sets, "stopDate=NULL")
		}
	}
	sets = append(sets, fmt.Sprintf("userModificationDate=%f", now))

	sql := fmt.Sprintf(
		`UPDATE TMChecklistItem SET %s WHERE uuid='%s' AND task='%s'`,
		strings.Join(sets, ", "),
		escapeSQLite(itemID),
		escapeSQLite(taskID),
	)

	if _, err := query(sql); err != nil {
		return nil, fmt.Errorf("failed to update checklist item: %w", err)
	}

	readSQL := fmt.Sprintf(
		`SELECT uuid, title, status FROM TMChecklistItem WHERE uuid='%s' AND task='%s'`,
		escapeSQLite(itemID),
		escapeSQLite(taskID),
	)
	out, err := query(readSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to read updated checklist item: %w", err)
	}
	items := parseChecklistItems(out)
	if len(items) == 0 {
		return nil, fmt.Errorf("checklist item not found")
	}
	return &items[0], nil
}

// DeleteChecklistItem removes a checklist item via SQLite.
func DeleteChecklistItem(taskID, itemID string) error {
	if err := models.ValidateThingsID(taskID); err != nil {
		return err
	}
	if err := models.ValidateThingsID(itemID); err != nil {
		return err
	}

	sql := fmt.Sprintf(
		`DELETE FROM TMChecklistItem WHERE uuid='%s' AND task='%s'`,
		escapeSQLite(itemID),
		escapeSQLite(taskID),
	)

	if _, err := query(sql); err != nil {
		return fmt.Errorf("failed to delete checklist item: %w", err)
	}
	return nil
}

// parseChecklistItems parses sqlite3 tab-delimited output into ChecklistItem structs.
func parseChecklistItems(output string) []models.ChecklistItem {
	var items []models.ChecklistItem
	if output == "" {
		return items
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}
		items = append(items, models.ChecklistItem{
			ID:        fields[0],
			Title:     fields[1],
			Completed: fields[2] == "3",
		})
	}
	return items
}

// escapeSQLite escapes a string for safe use in SQLite queries.
func escapeSQLite(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// coreDataTimestamp returns the current time as a Core Data timestamp
// (seconds since 2001-01-01 00:00:00 UTC).
func coreDataTimestamp() float64 {
	epoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	return time.Since(epoch).Seconds()
}

// generateUUID generates a Things-style UUID (22 chars, base62).
func generateUUID() string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 22)
	r, _ := exec.Command("dd", "if=/dev/urandom", "bs=22", "count=1", "status=none").Output()
	if len(r) < 22 {
		r = []byte(fmt.Sprintf("%022d", time.Now().UnixNano()))
	}
	for i := 0; i < 22; i++ {
		b[i] = chars[r[i]%62]
	}
	return string(b)
}
