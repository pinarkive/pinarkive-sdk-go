package pinarkive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// PinarkiveClient represents a client for the Pinarkive API
type PinarkiveClient struct {
	BaseURL string
	Token   string
	APIKey  string
	Client  *http.Client
}

// FileUpload represents a file to be uploaded with path and content
type FileUpload struct {
	Path    string
	Content io.Reader
}

// TokenOptions represents options for token generation
type TokenOptions struct {
	Permissions   []string `json:"permissions,omitempty"`
	ExpiresInDays int      `json:"expiresInDays,omitempty"`
	IPAllowlist   []string `json:"ipAllowlist,omitempty"`
}

// NewPinarkiveClient creates a new Pinarkive client
func NewPinarkiveClient(token, apiKey, baseURL string) *PinarkiveClient {
	if baseURL == "" {
		baseURL = "https://api.pinarkive.com/api/v2"
	}
	return &PinarkiveClient{
		BaseURL: baseURL,
		Token:   token,
		APIKey:  apiKey,
		Client:  &http.Client{},
	}
}

func (c *PinarkiveClient) headers() http.Header {
	headers := http.Header{}
	if c.Token != "" {
		headers.Set("Authorization", "Bearer "+c.Token)
	} else if c.APIKey != "" {
		headers.Set("Authorization", "Bearer "+c.APIKey)
	}
	return headers
}

// --- Authentication ---

// --- File Management ---

// UploadFile uploads a single file
func (c *PinarkiveClient) UploadFile(filePath string) (*http.Response, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return nil, err
	}
	w.Close()
	url := c.BaseURL + "/files"
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}
	req.Header = c.headers()
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.Client.Do(req)
}

// UploadDirectory uploads a directory from local path
func (c *PinarkiveClient) UploadDirectory(dirPath string) (*http.Response, error) {
	data := map[string]string{"dirPath": dirPath}
	return c.postJson("/files/directory", data, c.headers())
}

// UploadDirectoryDAG uploads directory structure as DAG (Directed Acyclic Graph)
func (c *PinarkiveClient) UploadDirectoryDAG(files map[string]io.Reader, dirName string) (*http.Response, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add directory name if provided
	if dirName != "" {
		w.WriteField("dirName", dirName)
	}

	// Add files
	index := 0
	for path, content := range files {
		// Add path
		pathField := fmt.Sprintf("files[%d][path]", index)
		w.WriteField(pathField, path)

		// Add content
		contentField := fmt.Sprintf("files[%d][content]", index)
		fw, err := w.CreateFormField(contentField)
		if err != nil {
			return nil, err
		}
		if _, err = io.Copy(fw, content); err != nil {
			return nil, err
		}
		index++
	}

	w.Close()
	url := c.BaseURL + "/files/directory-dag"
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}
	req.Header = c.headers()
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.Client.Do(req)
}

// RenameFile renames an uploaded file
func (c *PinarkiveClient) RenameFile(uploadID, newName string) (*http.Response, error) {
	data := map[string]string{"newName": newName}
	return c.putJson("/files/rename/"+uploadID, data, c.headers())
}

// PinCid pins a CID to your account
func (c *PinarkiveClient) PinCid(cid, filename string) (*http.Response, error) {
	data := map[string]string{}
	if filename != "" {
		data["filename"] = filename
	}
	return c.postJson("/files/pin/"+cid, data, c.headers())
}

// RemoveFile removes a file from storage
func (c *PinarkiveClient) RemoveFile(cid string) (*http.Response, error) {
	url := c.BaseURL + "/files/remove/" + cid
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.headers()
	return c.Client.Do(req)
}

// ListUploads lists uploaded files with pagination
func (c *PinarkiveClient) ListUploads(page, limit int) (*http.Response, error) {
	url := fmt.Sprintf("/users/me/uploads?page=%d&limit=%d", page, limit)
	return c.get(url, c.headers())
}

// --- Token Management ---

// GenerateToken generates an API token with enhanced options
func (c *PinarkiveClient) GenerateToken(name string, options TokenOptions) (*http.Response, error) {
	data := map[string]interface{}{"name": name}

	if len(options.Permissions) > 0 {
		data["permissions"] = options.Permissions
	}
	if options.ExpiresInDays > 0 {
		data["expiresInDays"] = options.ExpiresInDays
	}
	if len(options.IPAllowlist) > 0 {
		data["ipAllowlist"] = options.IPAllowlist
	}

	return c.postJson("/tokens/generate", data, c.headers())
}

// ListTokens lists all API tokens
func (c *PinarkiveClient) ListTokens() (*http.Response, error) {
	return c.get("/tokens/list", c.headers())
}

// RevokeToken revokes an API token
func (c *PinarkiveClient) RevokeToken(name string) (*http.Response, error) {
	url := "/tokens/revoke/" + name
	return c.delete(url, c.headers())
}

// --- Status and Monitoring ---

// GetStatus gets file status
func (c *PinarkiveClient) GetStatus(cid string) (*http.Response, error) {
	url := "/status/" + cid
	return c.get(url, c.headers())
}

// GetAllocations gets storage allocations for a CID
func (c *PinarkiveClient) GetAllocations(cid string) (*http.Response, error) {
	url := "/status/allocations/" + cid
	return c.get(url, c.headers())
}

// --- Helper HTTP methods ---

func (c *PinarkiveClient) get(path string, headers http.Header) (*http.Response, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	return c.Client.Do(req)
}

func (c *PinarkiveClient) postJson(path string, data interface{}, headers http.Header) (*http.Response, error) {
	url := c.BaseURL + path
	var body io.Reader
	if data != nil {
		b, _ := json.Marshal(data)
		body = bytes.NewBuffer(b)
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Client.Do(req)
}

func (c *PinarkiveClient) putJson(path string, data interface{}, headers http.Header) (*http.Response, error) {
	url := c.BaseURL + path
	b, _ := json.Marshal(data)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header = headers
	req.Header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}

func (c *PinarkiveClient) delete(path string, headers http.Header) (*http.Response, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	return c.Client.Do(req)
}
