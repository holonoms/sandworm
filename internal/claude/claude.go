// Package claude provides functionality for interacting with the Claude API.
package claude

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/umwelt-studio/sandworm/internal/config"
)

const (
	baseURL        = "https://api.claude.ai"
	sessionKey     = "session_key"
	documentID     = "document_id"
	organizationID = "organization_id"
	projectID      = "project_id"
)

// Client manages interactions with the Claude API, handling authentication,
// project management, and document operations.
type Client struct {
	config     *config.Section
	httpClient *http.Client
}

// Required configuration keys for the client to function
var requiredKeys = []string{sessionKey, organizationID, projectID}
var sessionKeyRegex = regexp.MustCompile(`^sessionKey=([^;]+)`)

// New creates a new Claude API client using the provided configuration section.
func New(cfg *config.Section) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// Minimum TLS version required by Claude; 403's without this.
					MinVersion: tls.VersionTLS12,
				},
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
	}
}

// MARK: Interface

// Setup initializes the client configuration, prompting for required values
// if they're not already set. It validates organization access and project
// selection.
func (c *Client) Setup(force bool) (bool, error) {
	// Handle session key setup
	if force || !c.config.Has(sessionKey) {
		fmt.Println("\nPlease go to https://claude.ai in your browser and copy your session key from the Cookie header.")
		fmt.Println("You can find this in your browser's developer tools under Network tab.")
		fmt.Println()
		fmt.Print("Enter your session key: ")
		var key string
		if _, err := fmt.Scanln(&key); err != nil {
			return false, fmt.Errorf("failed to read session key: %w", err)
		}
		c.config.Set(sessionKey, key)
		if err := c.config.Save(); err != nil {
			return false, err
		}
	}

	// Handle organization selection
	if force || !c.config.Has(organizationID) {
		orgs, err := c.listOrganizations()
		if err != nil {
			return false, err
		}

		if len(orgs) == 0 {
			fmt.Println("\nNo organizations found. Please create one at https://claude.ai")
			return false, nil
		}

		var org organization
		if len(orgs) == 1 {
			org = orgs[0]
			fmt.Printf("\nUsing organization: %s\n", org.Name)
		} else {
			fmt.Println("\nSelect an organization:")
			org = selectFromList(orgs)
		}

		c.config.Set(organizationID, org.ID)
		if err := c.config.Save(); err != nil {
			return false, err
		}
	}

	// Handle project selection
	if force || !c.config.Has(projectID) {
		projects, err := c.listProjects()
		if err != nil {
			return false, err
		}

		var activeProjects []project
		for _, p := range projects {
			if p.ArchivedAt.IsZero() {
				activeProjects = append(activeProjects, p)
			}
		}

		if len(activeProjects) == 0 {
			fmt.Println("\nNo active projects found. Please create one at https://claude.ai")
			return false, nil
		}

		fmt.Println("\nSelect a project:")
		project := selectFromList(activeProjects)
		c.config.Set(projectID, project.ID)
		if err := c.config.Save(); err != nil {
			return false, err
		}
	}

	return true, nil
}

// Push uploads a file to the selected Claude project. If a file with the same
// name exists, it's replaced.
func (c *Client) Push(filePath, fileName string) error {
	if err := c.validateConfig(); err != nil {
		return err
	}

	// If no document ID is set, try to find existing document
	if !c.config.Has(documentID) {
		docs, err := c.listDocuments()
		if err != nil {
			return err
		}
		for _, doc := range docs {
			if doc.FileName == fileName {
				c.config.Set(documentID, doc.ID)
				if err := c.config.Save(); err != nil {
					return err
				}
				break
			}
		}
	}

	// Delete existing document if we have one
	if c.config.Has(documentID) {
		if err := c.deleteDocument(c.config.Get(documentID)); err != nil {
			// Only return error if it's not a 404
			if !strings.Contains(err.Error(), "404") {
				return err
			}
		}
		c.config.Delete(documentID)
		if err := c.config.Save(); err != nil {
			return err
		}
	}

	// Read and upload new file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	doc, err := c.uploadDocument(fileName, string(content))
	if err != nil {
		return err
	}

	c.config.Set(documentID, doc.ID)
	return c.config.Save()
}

// PurgeProjectFiles removes all files from the current project.
func (c *Client) PurgeProjectFiles(progressFn func(fileName string, current, total int)) (int, error) {
	if err := c.validateConfig(); err != nil {
		return 0, err
	}

	docs, err := c.listDocuments()
	if err != nil {
		return 0, err
	}

	for i, doc := range docs {
		if progressFn != nil {
			progressFn(doc.FileName, i+1, len(docs))
		}

		if err := c.deleteDocument(doc.ID); err != nil {
			// Only return error if it's not a 404
			if !strings.Contains(err.Error(), "404") {
				return i, err
			}
		}
	}

	c.config.Delete(documentID)
	if err := c.config.Save(); err != nil {
		return len(docs), err
	}

	return len(docs), nil
}

// MARK: Internal helper function

// validateConfig ensures all required configuration values are present
func (c *Client) validateConfig() error {
	var missing []string
	for _, key := range requiredKeys {
		if !c.config.Has(key) {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config keys: %s", strings.Join(missing, ", "))
	}
	return nil
}

// makeRequest performs an HTTP request to the Claude API
func (c *Client) makeRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+"/api"+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:129.0) Gecko/20100101 Firefox/129.0",
		"Cookie":       fmt.Sprintf("sessionKey=%s", c.config.Get(sessionKey)),
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed: %d - %s", resp.StatusCode, string(respBody))
	}

	// Update session key if it changed
	if cookie := resp.Header.Get("Set-Cookie"); cookie != "" {
		if matches := sessionKeyRegex.FindStringSubmatch(cookie); matches != nil {
			newKey := matches[1]
			if newKey != c.config.Get(sessionKey) {
				c.config.Set(sessionKey, newKey)
				_ = c.config.Save()
			}
		}
	}

	return respBody, nil
}

// MARK: Anthropic API requests

func (c *Client) listOrganizations() ([]organization, error) {
	data, err := c.makeRequest(http.MethodGet, "/organizations", nil)
	if err != nil {
		return nil, fmt.Errorf("listOrganizations: %w", err)
	}

	var orgs []organization
	if err := json.Unmarshal(data, &orgs); err != nil {
		return nil, fmt.Errorf("failed to parse organizations: %w", err)
	}
	return orgs, nil
}

func (c *Client) listProjects() ([]project, error) {
	data, err := c.makeRequest(
		http.MethodGet,
		fmt.Sprintf("/organizations/%s/projects", c.config.Get(organizationID)),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("listProjects: %w", err)
	}

	var projects []project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects: %w", err)
	}
	return projects, nil
}

func (c *Client) listDocuments() ([]document, error) {
	data, err := c.makeRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/organizations/%s/projects/%s/docs",
			c.config.Get(organizationID),
			c.config.Get(projectID),
		),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("listDocuments: %w", err)
	}

	var docs []document
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("failed to parse documents: %w", err)
	}
	return docs, nil
}

func (c *Client) deleteDocument(id string) error {
	_, err := c.makeRequest(
		http.MethodDelete,
		fmt.Sprintf(
			"/organizations/%s/projects/%s/docs/%s",
			c.config.Get(organizationID),
			c.config.Get(projectID),
			id,
		),
		nil,
	)
	if err != nil {
		return fmt.Errorf("deleteDocument: %w", err)
	}
	return nil
}

func (c *Client) uploadDocument(fileName, content string) (*document, error) {
	body := map[string]string{
		"file_name": fileName,
		"content":   content,
	}

	data, err := c.makeRequest(
		http.MethodPost,
		fmt.Sprintf(
			"/organizations/%s/projects/%s/docs",
			c.config.Get(organizationID),
			c.config.Get(projectID),
		),
		body,
	)
	if err != nil {
		return nil, fmt.Errorf("uploadDocument: %w", err)
	}

	var doc document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}
	return &doc, nil
}

// MARK: Anthropic API types

type organization struct {
	ID   string `json:"uuid"`
	Name string `json:"name"`
}

type project struct {
	ID         string    `json:"uuid"`
	Name       string    `json:"name"`
	ArchivedAt time.Time `json:"archived_at,omitempty"`
}

type document struct {
	ID       string `json:"uuid"`
	FileName string `json:"file_name"`
}

// MARK: User interaction (for setup)

// Helper function to present a selection list to the user and return the
// selected item.
func selectFromList[T interface{ GetName() string }](items []T) T {
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item.GetName())
	}

	for {
		fmt.Print("\nEnter selection number: ")
		var input int
		if _, err := fmt.Scanln(&input); err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		if input < 1 || input > len(items) {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d\n", len(items))
			continue
		}

		return items[input-1]
	}
}

// GetName implementations for our types to satisfy the generic constraint
func (o organization) GetName() string { return o.Name }
func (p project) GetName() string      { return p.Name }
