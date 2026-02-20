package applescript

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egorkaBurkenya/things3-api/models"
)

// GetAllProjects retrieves all projects from Things 3.
func GetAllProjects() ([]models.Project, error) {
	script := `tell application "Things3"
	set output to ""
	repeat with p in projects
		set pId to id of p
		set pName to name of p
		set pNotes to notes of p
		set pArea to ""
		try
			set pArea to name of area of p
		end try
		set taskCount to count of to dos of p
		set output to output & pId & tab & pName & tab & pNotes & tab & pArea & tab & (taskCount as string) & linefeed
	end repeat
	return output
end tell`

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	return parseProjects(out), nil
}

// parseProjects parses tab-delimited AppleScript output into Project structs.
// Expected format per line: id<TAB>name<TAB>notes<TAB>area<TAB>taskCount
func parseProjects(output string) []models.Project {
	var projects []models.Project
	if output == "" {
		return projects
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}

		project := models.Project{
			ID:   fields[0],
			Name: fields[1],
		}

		if len(fields) > 2 {
			project.Notes = cleanMissingValue(fields[2])
		}
		if len(fields) > 3 {
			project.Area = cleanMissingValue(fields[3])
		}
		if len(fields) > 4 {
			count, err := strconv.Atoi(strings.TrimSpace(fields[4]))
			if err == nil {
				project.TaskCount = count
			}
		}

		projects = append(projects, project)
	}
	return projects
}

// GetProjectByID retrieves a single project by its Things 3 ID.
func GetProjectByID(id string) (*models.Project, error) {
	if err := models.ValidateThingsID(id); err != nil {
		return nil, err
	}

	script := fmt.Sprintf(`tell application "Things3"
	set p to first project whose id is "%s"
	set pId to id of p
	set pName to name of p
	set pNotes to notes of p
	set pArea to ""
	try
		set pArea to name of area of p
	end try
	set taskCount to count of to dos of p
	return pId & tab & pName & tab & pNotes & tab & pArea & tab & (taskCount as string)
end tell`, EscapeString(id))

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", id, err)
	}

	projects := parseProjects(out)
	if len(projects) == 0 {
		return nil, fmt.Errorf("project %s not found", id)
	}
	return &projects[0], nil
}

// CreateProject creates a new project in Things 3 and returns the created project.
func CreateProject(req models.CreateProjectRequest) (*models.Project, error) {
	props := fmt.Sprintf(`name:"%s"`, EscapeString(req.Name))
	if req.Notes != "" {
		props += fmt.Sprintf(`, notes:"%s"`, EscapeString(req.Notes))
	}

	var scriptParts []string
	scriptParts = append(scriptParts, `tell application "Things3"`)
	scriptParts = append(scriptParts,
		fmt.Sprintf(`	set newProj to make new project with properties {%s}`, props),
	)

	// Assign to area if specified.
	if req.Area != "" {
		scriptParts = append(scriptParts,
			FindByNameScript("area", "targetArea", req.Area),
			`	set area of newProj to targetArea`,
		)
	}

	// Set scheduling via the "when" field.
	if req.When != "" {
		switch req.When {
		case "today":
			scriptParts = append(scriptParts,
				`	move newProj to list "Today"`,
			)
		case "someday":
			scriptParts = append(scriptParts,
				`	move newProj to list "Someday"`,
			)
		case "anytime":
			scriptParts = append(scriptParts,
				`	move newProj to list "Anytime"`,
			)
		default:
			// Assume it's a date string (YYYY-MM-DD).
			scriptParts = append(scriptParts,
				fmt.Sprintf(`	set activation date of newProj to date "%s"`, EscapeString(req.When)),
			)
		}
	}

	scriptParts = append(scriptParts,
		`	return id of newProj`,
		`end tell`,
	)

	script := strings.Join(scriptParts, "\n")

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	newID := strings.TrimSpace(out)
	return GetProjectByID(newID)
}

// UpdateProject updates an existing project and returns the updated project.
func UpdateProject(id string, req models.UpdateProjectRequest) (*models.Project, error) {
	if err := models.ValidateThingsID(id); err != nil {
		return nil, err
	}

	var scriptParts []string
	scriptParts = append(scriptParts,
		`tell application "Things3"`,
		fmt.Sprintf(`	set p to first project whose id is "%s"`, EscapeString(id)),
	)

	if req.Name != nil {
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set name of p to "%s"`, EscapeString(*req.Name)),
		)
	}
	if req.Notes != nil {
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set notes of p to "%s"`, EscapeString(*req.Notes)),
		)
	}
	if req.Area != nil {
		if *req.Area == "" {
			scriptParts = append(scriptParts,
				`	set area of p to missing value`,
			)
		} else {
			scriptParts = append(scriptParts,
				FindByNameScript("area", "targetArea", *req.Area),
				`	set area of p to targetArea`,
			)
		}
	}

	scriptParts = append(scriptParts, `end tell`)

	script := strings.Join(scriptParts, "\n")

	_, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to update project %s: %w", id, err)
	}

	return GetProjectByID(id)
}

// CompleteProject marks a project as completed.
func CompleteProject(id string) error {
	if err := models.ValidateThingsID(id); err != nil {
		return err
	}

	script := fmt.Sprintf(`tell application "Things3"
	set p to first project whose id is "%s"
	set status of p to completed
end tell`, EscapeString(id))

	_, err := Run(script)
	if err != nil {
		return fmt.Errorf("failed to complete project %s: %w", id, err)
	}
	return nil
}

// GetAllAreas retrieves all areas from Things 3.
func GetAllAreas() ([]models.Area, error) {
	script := `tell application "Things3"
	set output to ""
	repeat with a in areas
		set aId to id of a
		set aName to name of a
		set output to output & aId & tab & aName & linefeed
	end repeat
	return output
end tell`

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get areas: %w", err)
	}
	return parseAreas(out), nil
}

// parseAreas parses tab-delimited AppleScript output into Area structs.
// Expected format per line: id<TAB>name
func parseAreas(output string) []models.Area {
	var areas []models.Area
	if output == "" {
		return areas
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}

		area := models.Area{
			ID:   fields[0],
			Name: fields[1],
		}

		areas = append(areas, area)
	}
	return areas
}

// GetAreaByID retrieves a single area by its Things 3 ID.
func GetAreaByID(id string) (*models.Area, error) {
	if err := models.ValidateThingsID(id); err != nil {
		return nil, err
	}

	script := fmt.Sprintf(`tell application "Things3"
	set a to first area whose id is "%s"
	set aId to id of a
	set aName to name of a
	return aId & tab & aName
end tell`, EscapeString(id))

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get area %s: %w", id, err)
	}

	areas := parseAreas(out)
	if len(areas) == 0 {
		return nil, fmt.Errorf("area %s not found", id)
	}

	// Fetch projects belonging to this area.
	projects, err := getProjectsForArea(id)
	if err == nil {
		areas[0].Projects = projects
	}

	return &areas[0], nil
}

// getProjectsForArea retrieves all projects belonging to a specific area.
func getProjectsForArea(areaID string) ([]models.Project, error) {
	script := fmt.Sprintf(`tell application "Things3"
	set a to first area whose id is "%s"
	set output to ""
	repeat with p in projects of a
		set pId to id of p
		set pName to name of p
		set pNotes to notes of p
		set taskCount to count of to dos of p
		set output to output & pId & tab & pName & tab & pNotes & tab & "" & tab & (taskCount as string) & linefeed
	end repeat
	return output
end tell`, EscapeString(areaID))

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects for area %s: %w", areaID, err)
	}
	return parseProjects(out), nil
}

// CreateArea creates a new area in Things 3 and returns the created area.
func CreateArea(req models.CreateAreaRequest) (*models.Area, error) {
	script := fmt.Sprintf(`tell application "Things3"
	set newArea to make new area with properties {name:"%s"}
	return id of newArea
end tell`, EscapeString(req.Name))

	out, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to create area: %w", err)
	}

	newID := strings.TrimSpace(out)
	return GetAreaByID(newID)
}

// UpdateArea updates an existing area and returns the updated area.
func UpdateArea(id string, req models.UpdateAreaRequest) (*models.Area, error) {
	if err := models.ValidateThingsID(id); err != nil {
		return nil, err
	}

	var scriptParts []string
	scriptParts = append(scriptParts,
		`tell application "Things3"`,
		fmt.Sprintf(`	set a to first area whose id is "%s"`, EscapeString(id)),
	)

	if req.Name != nil {
		scriptParts = append(scriptParts,
			fmt.Sprintf(`	set name of a to "%s"`, EscapeString(*req.Name)),
		)
	}

	scriptParts = append(scriptParts, `end tell`)

	script := strings.Join(scriptParts, "\n")

	_, err := Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to update area %s: %w", id, err)
	}

	return GetAreaByID(id)
}
