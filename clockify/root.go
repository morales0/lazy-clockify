package clockify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.clockify.me/api/v1"

// Client represents a Clockify API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Clockify API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// doRequest performs an HTTP request with the API key header
func (c *Client) doRequest(method, path string, body any) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// User represents a Clockify user
type User struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	DefaultWorkspace string `json:"defaultWorkspace"`
}

// Workspace represents a Clockify workspace
type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Project represents a Clockify project
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TimeEntryRequest represents a request to create a time entry
type TimeEntryRequest struct {
	Start       time.Time  `json:"start"`
	End         *time.Time `json:"end,omitempty"`
	Billable    bool       `json:"billable"`
	Description string     `json:"description"`
	ProjectID   *string    `json:"projectId"`
	TaskID      *string    `json:"taskId,omitempty"`
	TagIDs      []string   `json:"tagIds,omitempty"`
}

// TimeEntry represents a Clockify time entry response
type TimeEntry struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Start       time.Time `json:"timeInterval.start"`
	End         time.Time `json:"timeInterval.end"`
	WorkspaceID string    `json:"workspaceId"`
	UserID      string    `json:"userId"`
}

// GetUser retrieves the current user information
func (c *Client) GetUser() (*User, error) {
	resp, err := c.doRequest("GET", "/user", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// GetWorkspaces retrieves all workspaces for the user
func (c *Client) GetWorkspaces() ([]Workspace, error) {
	resp, err := c.doRequest("GET", "/workspaces", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var workspaces []Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspaces); err != nil {
		return nil, fmt.Errorf("failed to decode workspaces response: %w", err)
	}

	return workspaces, nil
}

// GetProjects retrieves all projects for a workspace
func (c *Client) GetProjects(workspaceID string) ([]Project, error) {
	path := fmt.Sprintf("/workspaces/%s/projects", workspaceID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode projects response: %w", err)
	}

	return projects, nil
}

// CreateTimeEntry creates a new time entry in the specified workspace
func (c *Client) CreateTimeEntry(workspaceID string, entry TimeEntryRequest) (*TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%s/time-entries", workspaceID)

	resp, err := c.doRequest("POST", path, entry)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var timeEntry TimeEntry
	if err := json.NewDecoder(resp.Body).Decode(&timeEntry); err != nil {
		return nil, fmt.Errorf("failed to decode time entry response: %w", err)
	}

	return &timeEntry, nil
}
