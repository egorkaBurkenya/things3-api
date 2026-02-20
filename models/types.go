package models

import (
	"fmt"
	"regexp"
	"time"
)

type Task struct {
	ID             string          `json:"id"`
	Title          string          `json:"title"`
	Notes          string          `json:"notes,omitempty"`
	Status         string          `json:"status"`
	Project        string          `json:"project,omitempty"`
	Area           string          `json:"area,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
	Due            string          `json:"due,omitempty"`
	When           string          `json:"when,omitempty"`
	CreatedAt      string          `json:"created_at,omitempty"`
	ChecklistItems []ChecklistItem `json:"checklist_items,omitempty"`
}

type ChecklistItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type CreateTaskRequest struct {
	Title          string   `json:"title"`
	Notes          string   `json:"notes"`
	Project        string   `json:"project"`
	Area           string   `json:"area"`
	Due            string   `json:"due"`
	When           string   `json:"when"`
	Tags           []string `json:"tags"`
	ChecklistItems []string `json:"checklistItems"`
}

func (r *CreateTaskRequest) Validate() error {
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(r.Title) > 1000 {
		return fmt.Errorf("title must be under 1000 characters")
	}
	if len(r.Notes) > 10000 {
		return fmt.Errorf("notes must be under 10000 characters")
	}
	if r.Due != "" {
		if _, err := time.Parse("2006-01-02", r.Due); err != nil {
			return fmt.Errorf("due must be ISO 8601 date (YYYY-MM-DD)")
		}
	}
	if r.When != "" && !isValidWhen(r.When) {
		return fmt.Errorf("when must be one of: today, evening, tomorrow, someday, anytime, or a date (YYYY-MM-DD)")
	}
	if len(r.Tags) > 50 {
		return fmt.Errorf("maximum 50 tags allowed")
	}
	for _, tag := range r.Tags {
		if len(tag) > 200 {
			return fmt.Errorf("each tag must be under 200 characters")
		}
	}
	if len(r.Project) > 500 {
		return fmt.Errorf("project name must be under 500 characters")
	}
	if len(r.Area) > 500 {
		return fmt.Errorf("area name must be under 500 characters")
	}
	if len(r.ChecklistItems) > 100 {
		return fmt.Errorf("maximum 100 checklist items allowed")
	}
	for _, item := range r.ChecklistItems {
		if item == "" {
			return fmt.Errorf("checklist item title cannot be empty")
		}
		if len(item) > 1000 {
			return fmt.Errorf("checklist item title must be under 1000 characters")
		}
	}
	return nil
}

type UpdateTaskRequest struct {
	Title   *string  `json:"title"`
	Notes   *string  `json:"notes"`
	Project *string  `json:"project"`
	Area    *string  `json:"area"`
	Due     *string  `json:"due"`
	When    *string  `json:"when"`
	Tags    []string `json:"tags"`
}

func (r *UpdateTaskRequest) Validate() error {
	if r.Title != nil {
		if *r.Title == "" {
			return fmt.Errorf("title cannot be empty")
		}
		if len(*r.Title) > 1000 {
			return fmt.Errorf("title must be under 1000 characters")
		}
	}
	if r.Notes != nil && len(*r.Notes) > 10000 {
		return fmt.Errorf("notes must be under 10000 characters")
	}
	if r.Due != nil && *r.Due != "" {
		if _, err := time.Parse("2006-01-02", *r.Due); err != nil {
			return fmt.Errorf("due must be ISO 8601 date (YYYY-MM-DD)")
		}
	}
	if r.When != nil && *r.When != "" && !isValidWhen(*r.When) {
		return fmt.Errorf("when must be one of: today, evening, tomorrow, someday, anytime, or a date (YYYY-MM-DD)")
	}
	if r.Tags != nil {
		if len(r.Tags) > 50 {
			return fmt.Errorf("maximum 50 tags allowed")
		}
		for _, tag := range r.Tags {
			if len(tag) > 200 {
				return fmt.Errorf("each tag must be under 200 characters")
			}
		}
	}
	if r.Project != nil && len(*r.Project) > 500 {
		return fmt.Errorf("project name must be under 500 characters")
	}
	if r.Area != nil && len(*r.Area) > 500 {
		return fmt.Errorf("area name must be under 500 characters")
	}
	return nil
}

type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Area      string `json:"area,omitempty"`
	Notes     string `json:"notes,omitempty"`
	TaskCount int    `json:"task_count,omitempty"`
}

type CreateProjectRequest struct {
	Name  string `json:"name"`
	Area  string `json:"area"`
	Notes string `json:"notes"`
	When  string `json:"when"`
}

func (r *CreateProjectRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Name) > 500 {
		return fmt.Errorf("name must be under 500 characters")
	}
	if len(r.Notes) > 10000 {
		return fmt.Errorf("notes must be under 10000 characters")
	}
	if len(r.Area) > 500 {
		return fmt.Errorf("area name must be under 500 characters")
	}
	if r.When != "" && !isValidProjectWhen(r.When) {
		return fmt.Errorf("when must be one of: today, someday, anytime, or a date (YYYY-MM-DD)")
	}
	return nil
}

type UpdateProjectRequest struct {
	Name  *string `json:"name"`
	Area  *string `json:"area"`
	Notes *string `json:"notes"`
}

func (r *UpdateProjectRequest) Validate() error {
	if r.Name != nil {
		if *r.Name == "" {
			return fmt.Errorf("name cannot be empty")
		}
		if len(*r.Name) > 500 {
			return fmt.Errorf("name must be under 500 characters")
		}
	}
	if r.Notes != nil && len(*r.Notes) > 10000 {
		return fmt.Errorf("notes must be under 10000 characters")
	}
	if r.Area != nil && len(*r.Area) > 500 {
		return fmt.Errorf("area name must be under 500 characters")
	}
	return nil
}

type Area struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Projects []Project `json:"projects,omitempty"`
}

type CreateAreaRequest struct {
	Name string `json:"name"`
}

func (r *CreateAreaRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Name) > 500 {
		return fmt.Errorf("name must be under 500 characters")
	}
	return nil
}

type UpdateAreaRequest struct {
	Name *string `json:"name"`
}

func (r *UpdateAreaRequest) Validate() error {
	if r.Name != nil {
		if *r.Name == "" {
			return fmt.Errorf("name cannot be empty")
		}
		if len(*r.Name) > 500 {
			return fmt.Errorf("name must be under 500 characters")
		}
	}
	return nil
}

type CreateChecklistItemRequest struct {
	Title string `json:"title"`
}

func (r *CreateChecklistItemRequest) Validate() error {
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(r.Title) > 1000 {
		return fmt.Errorf("title must be under 1000 characters")
	}
	return nil
}

type UpdateChecklistItemRequest struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

func (r *UpdateChecklistItemRequest) Validate() error {
	if r.Title != nil {
		if *r.Title == "" {
			return fmt.Errorf("title cannot be empty")
		}
		if len(*r.Title) > 1000 {
			return fmt.Errorf("title must be under 1000 characters")
		}
	}
	if r.Title == nil && r.Completed == nil {
		return fmt.Errorf("at least one field (title or completed) is required")
	}
	return nil
}

var thingsIDPattern = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)

func ValidateThingsID(id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if len(id) > 100 {
		return fmt.Errorf("invalid id format")
	}
	if !thingsIDPattern.MatchString(id) {
		return fmt.Errorf("invalid id format")
	}
	return nil
}

func isValidWhen(w string) bool {
	switch w {
	case "today", "evening", "tomorrow", "someday", "anytime":
		return true
	default:
		_, err := time.Parse("2006-01-02", w)
		return err == nil
	}
}

func isValidProjectWhen(w string) bool {
	switch w {
	case "today", "someday", "anytime":
		return true
	default:
		_, err := time.Parse("2006-01-02", w)
		return err == nil
	}
}
