package applescript

import (
	"fmt"
	"strings"

	"github.com/egorkaBurkenya/things3-api/models"
)

// getTasksFromList retrieves all tasks from a named Things 3 list.
func getTasksFromList(listName string) ([]models.Task, error) {
	script := fmt.Sprintf(`tell application "Things3"
	set taskList to to dos of list "%s"
	set output to ""
	repeat with t in taskList
		set taskId to id of t
		set taskName to name of t
		set taskNotes to notes of t
		set taskStatus to status of t
		set projName to ""
		try
			set projName to name of project of t
		end try
		set areaName to ""
		try
			set areaName to name of area of t
		end try
		set tagList to ""
		try
			set tagList to tag names of t
		end try
		set dueVal to "missing value"
		try
			set dueVal to due date of t as string
		end try
		set createdVal to "missing value"
		try
			set createdVal to creation date of t as string
		end try
		set output to output & taskId & tab & taskName & tab & taskNotes & tab & (taskStatus as string) & tab & projName & tab & areaName & tab & tagList & tab & dueVal & tab & createdVal & linefeed
	end repeat
	return output
end tell`, EscapeString(listName))

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks from %s: %w", listName, err)
	}
	return parseTasks(out), nil
}

// parseTasks parses tab-delimited AppleScript output into Task structs.
// Expected format per line: id<TAB>name<TAB>notes<TAB>status<TAB>project<TAB>area<TAB>tags<TAB>dueDate<TAB>createdDate
func parseTasks(output string) []models.Task {
	var tasks []models.Task
	if output == "" {
		return tasks
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			continue
		}

		task := models.Task{
			ID:     fields[0],
			Title:  fields[1],
			Notes:  cleanMissingValue(fields[2]),
			Status: normalizeStatus(fields[3]),
		}

		if len(fields) > 4 {
			task.Project = cleanMissingValue(fields[4])
		}
		if len(fields) > 5 {
			task.Area = cleanMissingValue(fields[5])
		}
		if len(fields) > 6 {
			task.Tags = parseTags(fields[6])
		}
		if len(fields) > 7 {
			task.Due = cleanMissingValue(fields[7])
		}
		if len(fields) > 8 {
			task.CreatedAt = cleanMissingValue(fields[8])
		}

		tasks = append(tasks, task)
	}
	return tasks
}

// parseTask parses a single task from tab-delimited output.
func parseTask(output string) *models.Task {
	tasks := parseTasks(output)
	if len(tasks) == 0 {
		return nil
	}
	return &tasks[0]
}

// cleanMissingValue converts AppleScript's "missing value" to an empty string.
func cleanMissingValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "missing value" {
		return ""
	}
	return s
}

// normalizeStatus converts AppleScript task status values to API-friendly strings.
func normalizeStatus(s string) string {
	s = strings.TrimSpace(s)
	switch s {
	case "open":
		return "open"
	case "completed":
		return "completed"
	case "canceled", "cancelled":
		return "canceled"
	default:
		return s
	}
}

// parseTags splits a comma-separated tag string into a slice.
func parseTags(s string) []string {
	s = cleanMissingValue(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ", ")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

// GetInboxTasks returns all tasks in the Inbox list.
func GetInboxTasks() ([]models.Task, error) {
	return getTasksFromList("Inbox")
}

// GetTodayTasks returns all tasks in the Today list.
func GetTodayTasks() ([]models.Task, error) {
	return getTasksFromList("Today")
}

// GetUpcomingTasks returns all tasks in the Upcoming list.
func GetUpcomingTasks() ([]models.Task, error) {
	return getTasksFromList("Upcoming")
}

// GetAnytimeTasks returns all tasks in the Anytime list.
func GetAnytimeTasks() ([]models.Task, error) {
	return getTasksFromList("Anytime")
}

// GetSomedayTasks returns all tasks in the Someday list.
func GetSomedayTasks() ([]models.Task, error) {
	return getTasksFromList("Someday")
}

// GetFilteredTasks retrieves tasks filtered by project, area, and/or tag.
// All filter parameters are optional; pass empty strings to skip a filter.
func GetFilteredTasks(project, area, tag string) ([]models.Task, error) {
	var script string

	switch {
	case project != "":
		findScript := FindByNameScript("project", "proj", project)
		script = fmt.Sprintf(`tell application "Things3"
%s
	set taskList to to dos of proj
	set output to ""
	repeat with t in taskList
		set taskId to id of t
		set taskName to name of t
		set taskNotes to notes of t
		set taskStatus to status of t
		set projName to ""
		try
			set projName to name of project of t
		end try
		set areaName to ""
		try
			set areaName to name of area of t
		end try
		set tagList to ""
		try
			set tagList to tag names of t
		end try
		set dueVal to "missing value"
		try
			set dueVal to due date of t as string
		end try
		set createdVal to "missing value"
		try
			set createdVal to creation date of t as string
		end try
		set output to output & taskId & tab & taskName & tab & taskNotes & tab & (taskStatus as string) & tab & projName & tab & areaName & tab & tagList & tab & dueVal & tab & createdVal & linefeed
	end repeat
	return output
end tell`, findScript)

	case area != "":
		findScript := FindByNameScript("area", "a", area)
		script = fmt.Sprintf(`tell application "Things3"
%s
	set taskList to to dos of a
	set output to ""
	repeat with t in taskList
		set taskId to id of t
		set taskName to name of t
		set taskNotes to notes of t
		set taskStatus to status of t
		set projName to ""
		try
			set projName to name of project of t
		end try
		set areaName to ""
		try
			set areaName to name of area of t
		end try
		set tagList to ""
		try
			set tagList to tag names of t
		end try
		set dueVal to "missing value"
		try
			set dueVal to due date of t as string
		end try
		set createdVal to "missing value"
		try
			set createdVal to creation date of t as string
		end try
		set output to output & taskId & tab & taskName & tab & taskNotes & tab & (taskStatus as string) & tab & projName & tab & areaName & tab & tagList & tab & dueVal & tab & createdVal & linefeed
	end repeat
	return output
end tell`, findScript)

	case tag != "":
		script = fmt.Sprintf(`tell application "Things3"
	set taskList to to dos whose tag names contains "%s"
	set output to ""
	repeat with t in taskList
		set taskId to id of t
		set taskName to name of t
		set taskNotes to notes of t
		set taskStatus to status of t
		set projName to ""
		try
			set projName to name of project of t
		end try
		set areaName to ""
		try
			set areaName to name of area of t
		end try
		set tagList to ""
		try
			set tagList to tag names of t
		end try
		set dueVal to "missing value"
		try
			set dueVal to due date of t as string
		end try
		set createdVal to "missing value"
		try
			set createdVal to creation date of t as string
		end try
		set output to output & taskId & tab & taskName & tab & taskNotes & tab & (taskStatus as string) & tab & projName & tab & areaName & tab & tagList & tab & dueVal & tab & createdVal & linefeed
	end repeat
	return output
end tell`, EscapeString(tag))

	default:
		return nil, fmt.Errorf("at least one filter (project, area, or tag) is required")
	}

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered tasks: %w", err)
	}
	return parseTasks(out), nil
}

// GetTaskByID retrieves a single task by its Things 3 ID.
func GetTaskByID(id string) (*models.Task, error) {
	if err := models.ValidateThingsID(id); err != nil {
		return nil, err
	}

	script := fmt.Sprintf(`tell application "Things3"
	set t to first to do whose id is "%s"
	set taskId to id of t
	set taskName to name of t
	set taskNotes to notes of t
	set taskStatus to status of t
	set projName to ""
	try
		set projName to name of project of t
	end try
	set areaName to ""
	try
		set areaName to name of area of t
	end try
	set tagList to ""
	try
		set tagList to tag names of t
	end try
	set dueVal to "missing value"
	try
		set dueVal to due date of t as string
	end try
	set createdVal to "missing value"
	try
		set createdVal to creation date of t as string
	end try
	return taskId & tab & taskName & tab & taskNotes & tab & (taskStatus as string) & tab & projName & tab & areaName & tab & tagList & tab & dueVal & tab & createdVal
end tell`, EscapeString(id))

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get task %s: %w", id, err)
	}

	task := parseTask(out)
	if task == nil {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return task, nil
}

// CreateTask creates a new task in Things 3 and returns the created task.
func CreateTask(req models.CreateTaskRequest) (*models.Task, error) {
	// Build the properties portion of the AppleScript.
	props := fmt.Sprintf(`name:"%s"`, EscapeString(req.Title))
	if req.Notes != "" {
		props += fmt.Sprintf(`, notes:"%s"`, EscapeString(req.Notes))
	}

	var scriptParts []string

	scriptParts = append(scriptParts, `tell application "Things3"`)

	// If a project is specified, create the task inside that project.
	if req.Project != "" {
		scriptParts = append(scriptParts,
			FindByNameScript("project", "proj", req.Project),
			fmt.Sprintf(`	set newTask to make new to do in proj with properties {%s}`, props),
		)
	} else {
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set newTask to make new to do with properties {%s}`, props),
		)
	}

	// Set due date if provided.
	if req.Due != "" {
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set due date of newTask to date "%s"`, EscapeString(req.Due)),
		)
	}

	// Set activation date / scheduling via the "when" field.
	if req.When != "" {
		switch req.When {
		case "today":
			scriptParts = append(scriptParts,
				`	move newTask to list "Today"`,
			)
		case "evening":
			scriptParts = append(scriptParts,
				`	move newTask to list "Today"`,
				`	set activation date of newTask to current date`,
			)
		case "someday":
			scriptParts = append(scriptParts,
				`	move newTask to list "Someday"`,
			)
		case "anytime":
			scriptParts = append(scriptParts,
				`	move newTask to list "Anytime"`,
			)
		case "tomorrow":
			scriptParts = append(scriptParts,
				`	set activation date of newTask to (current date) + 1 * days`,
			)
		default:
			// Assume it's a date string (YYYY-MM-DD).
			scriptParts = append(scriptParts,
				fmt.Sprintf(`	set activation date of newTask to date "%s"`, EscapeString(req.When)),
			)
		}
	}

	// Set tags if provided.
	if len(req.Tags) > 0 {
		escapedTags := make([]string, len(req.Tags))
		for i, t := range req.Tags {
			escapedTags[i] = EscapeString(t)
		}
		tagStr := strings.Join(escapedTags, ", ")
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set tag names of newTask to "%s"`, tagStr),
		)
	}

	// Move to area if specified (and no project given, since project takes precedence).
	if req.Area != "" && req.Project == "" {
		scriptParts = append(scriptParts,
			FindByNameScript("area", "targetArea", req.Area),
			`	set area of newTask to targetArea`,
		)
	}

	scriptParts = append(scriptParts,
		`	return id of newTask`,
		`end tell`,
	)

	script := strings.Join(scriptParts, "\n")

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	newID := strings.TrimSpace(out)
	return GetTaskByID(newID)
}

// UpdateTask updates an existing task and returns the updated task.
func UpdateTask(id string, req models.UpdateTaskRequest) (*models.Task, error) {
	if err := models.ValidateThingsID(id); err != nil {
		return nil, err
	}

	var scriptParts []string
	scriptParts = append(scriptParts,
		`tell application "Things3"`,
		fmt.Sprintf(`	set t to first to do whose id is "%s"`, EscapeString(id)),
	)

	if req.Title != nil {
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set name of t to "%s"`, EscapeString(*req.Title)),
		)
	}
	if req.Notes != nil {
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set notes of t to "%s"`, EscapeString(*req.Notes)),
		)
	}
	if req.Due != nil {
		if *req.Due == "" {
			scriptParts = append(scriptParts,
				`	set due date of t to missing value`,
			)
		} else {
			scriptParts = append(scriptParts,
				fmt.Sprintf(`	set due date of t to date "%s"`, EscapeString(*req.Due)),
			)
		}
	}
	if req.When != nil {
		when := *req.When
		switch when {
		case "today":
			scriptParts = append(scriptParts,
				`	move t to list "Today"`,
			)
		case "evening":
			scriptParts = append(scriptParts,
				`	move t to list "Today"`,
			)
		case "someday":
			scriptParts = append(scriptParts,
				`	move t to list "Someday"`,
			)
		case "anytime":
			scriptParts = append(scriptParts,
				`	move t to list "Anytime"`,
			)
		case "tomorrow":
			scriptParts = append(scriptParts,
				`	set activation date of t to (current date) + 1 * days`,
			)
		case "":
			scriptParts = append(scriptParts,
				`	set activation date of t to missing value`,
			)
		default:
			scriptParts = append(scriptParts,
				fmt.Sprintf(`	set activation date of t to date "%s"`, EscapeString(when)),
			)
		}
	}
	if req.Tags != nil {
		escapedTags := make([]string, len(req.Tags))
		for i, tag := range req.Tags {
			escapedTags[i] = EscapeString(tag)
		}
		tagStr := strings.Join(escapedTags, ", ")
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set tag names of t to "%s"`, tagStr),
		)
	}
	if req.Project != nil {
		if *req.Project == "" {
			scriptParts = append(scriptParts,
				`	move t to list "Inbox"`,
			)
		} else {
			scriptParts = append(scriptParts,
				FindByNameScript("project", "proj", *req.Project),
				`	move t to proj`,
			)
		}
	}

	scriptParts = append(scriptParts, `end tell`)

	script := strings.Join(scriptParts, "\n")

	_, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to update task %s: %w", id, err)
	}

	return GetTaskByID(id)
}

// CompleteTask marks a task as completed.
func CompleteTask(id string) error {
	if err := models.ValidateThingsID(id); err != nil {
		return err
	}

	script := fmt.Sprintf(`tell application "Things3"
	set t to first to do whose id is "%s"
	set status of t to completed
end tell`, EscapeString(id))

	_, err := Run(script)
	if err != nil {
		return fmt.Errorf("failed to complete task %s: %w", id, err)
	}
	return nil
}

// CancelTask marks a task as canceled.
func CancelTask(id string) error {
	if err := models.ValidateThingsID(id); err != nil {
		return err
	}

	script := fmt.Sprintf(`tell application "Things3"
	set t to first to do whose id is "%s"
	set status of t to canceled
end tell`, EscapeString(id))

	_, err := Run(script)
	if err != nil {
		return fmt.Errorf("failed to cancel task %s: %w", id, err)
	}
	return nil
}

// DeleteTask moves a task to the Trash list.
func DeleteTask(id string) error {
	if err := models.ValidateThingsID(id); err != nil {
		return err
	}

	script := fmt.Sprintf(`tell application "Things3"
	move (first to do whose id is "%s") to list "Trash"
end tell`, EscapeString(id))

	_, err := Run(script)
	if err != nil {
		return fmt.Errorf("failed to delete task %s: %w", id, err)
	}
	return nil
}

