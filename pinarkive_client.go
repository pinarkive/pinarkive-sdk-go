// Package pinarkive provides a minimal client for PinArkive API v3.
// See https://docs.pinarkive.com (upload, pin, remove, users/me, uploads, tokens, status, allocations).
// Auth: Bearer token or X-API-Key. On HTTP 4xx/5xx methods return *APIError with status code and API body.
//
// SDK version: 3.1.2
package pinarkive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Version is the SDK version (API v3).
const Version = "3.1.2"

// APIError is returned when the API responds with HTTP 4xx or 5xx.
// API v3 codes: 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found,
// 409 Conflict, 413 Payload Too Large, 429 Too Many Requests, 500 Internal Server Error, 503 Service Unavailable.
// For 403 missing_scope, check Required. For 429, check RetryAfterSeconds.
type APIError struct {
	StatusCode        int    // HTTP status (400, 401, 403, 404, 409, 413, 429, 500, 503)
	Err               string // API JSON field "error"
	Message           string // API JSON field "message"
	Code              string // API JSON field "code" (e.g. email_not_verified, missing_scope)
	Required          string // For 403 missing_scope: the required scope
	RetryAfterSeconds int    // For 429: seconds until retry (0 if not set)
	Body              []byte // Raw response body
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("pinarkive api [%d]: %s", e.StatusCode, e.Message)
	}
	if e.Err != "" {
		return fmt.Sprintf("pinarkive api [%d]: %s", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("pinarkive api [%d]", e.StatusCode)
}

// Client for PinArkive API v3.
type Client struct {
	BaseURL           string
	Token             string
	APIKey            string
	HTTP              *http.Client
	// RequestSourceWeb, when true, sends X-Request-Source: web on Bearer-authenticated requests only (not when using API Key). Set this when using the client from a web app so the backend classifies requests as WEB in logs.
	RequestSourceWeb bool
}

// NewClient creates a client. baseURL defaults to https://api.pinarkive.com/api/v3.
func NewClient(token, apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.pinarkive.com/api/v3"
	}
	if baseURL != "" && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		APIKey:  apiKey,
		HTTP:    &http.Client{},
	}
}

func (c *Client) headers(auth bool) http.Header {
	h := http.Header{}
	if auth {
		if c.APIKey != "" {
			h.Set("X-API-Key", c.APIKey)
		} else if c.Token != "" {
			h.Set("Authorization", "Bearer "+c.Token)
			if c.RequestSourceWeb {
				h.Set("X-Request-Source", "web")
			}
		}
	}
	return h
}

// --- Public (no auth) ---

// Health GET /health
func (c *Client) Health() (*http.Response, error) {
	return c.do("GET", "/health", nil, nil, false)
}

// GetPlans GET /plans/
func (c *Client) GetPlans() (*http.Response, error) {
	return c.do("GET", "/plans/", nil, nil, false)
}

// GetPeers GET /peers/
func (c *Client) GetPeers() (*http.Response, error) {
	return c.do("GET", "/peers/", nil, nil, false)
}

// Login POST /auth/login – body email, password. If response has requires2FA and temporaryToken, call Verify2FALogin.
func (c *Client) Login(email, password string) (*http.Response, error) {
	body := map[string]string{"email": email, "password": password}
	return c.do("POST", "/auth/login", body, nil, false)
}

// Verify2FALogin POST /auth/2fa/verify-login – complete login after 2FA; returns response with token.
func (c *Client) Verify2FALogin(temporaryToken, code string) (*http.Response, error) {
	body := map[string]string{"temporaryToken": temporaryToken, "code": code}
	return c.do("POST", "/auth/2fa/verify-login", body, nil, false)
}

// --- Files ---

// UploadFile POST /files/ – multipart file, optional cl, timelock (ISO 8601, premium)
func (c *Client) UploadFile(filePath string, clusterID, timelock *string) (*http.Response, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return nil, err
	}
	if clusterID != nil && *clusterID != "" {
		_ = w.WriteField("cl", *clusterID)
	}
	if timelock != nil && *timelock != "" {
		_ = w.WriteField("timelock", *timelock)
	}
	w.Close()
	return c.doMultipart("POST", "/files/", &buf, w.FormDataContentType())
}

// UploadDirectory POST /files/directory – body dirPath, optional cl, timelock
func (c *Client) UploadDirectory(dirPath string, clusterID, timelock *string) (*http.Response, error) {
	data := map[string]string{"dirPath": dirPath}
	if clusterID != nil && *clusterID != "" {
		data["cl"] = *clusterID
	}
	if timelock != nil && *timelock != "" {
		data["timelock"] = *timelock
	}
	return c.do("POST", "/files/directory", data, nil, true)
}

// UploadDirectoryDAG POST /files/directory-dag – multipart: repeated field name `files`, filename = path inside DAG; optional cl, timelock
func (c *Client) UploadDirectoryDAG(files map[string]io.Reader, dirName string, clusterID, timelock *string) (*http.Response, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if dirName != "" {
		_ = w.WriteField("dirName", dirName)
	}
	if clusterID != nil && *clusterID != "" {
		_ = w.WriteField("cl", *clusterID)
	}
	if timelock != nil && *timelock != "" {
		_ = w.WriteField("timelock", *timelock)
	}
	for p, content := range files {
		relPath := filepath.ToSlash(p)
		if relPath == "" || relPath[0] == '/' || strings.Contains(relPath, "..") {
			return nil, fmt.Errorf("invalid DAG path: %s", p)
		}
		// Repeated field name `files`, filename is the path inside the DAG.
		fw, err := w.CreateFormFile("files", relPath)
		if err != nil {
			return nil, err
		}
		if _, err = io.Copy(fw, content); err != nil {
			return nil, err
		}
	}
	w.Close()
	return c.doMultipart("POST", "/files/directory-dag", &buf, w.FormDataContentType())
}

// PinCid POST /files/pin/:cid – optional originalName, customName, cl, timelock
func (c *Client) PinCid(cid string, opts *PinOptions) (*http.Response, error) {
	data := map[string]string{}
	if opts != nil {
		if opts.OriginalName != "" {
			data["originalName"] = opts.OriginalName
		}
		if opts.CustomName != "" {
			data["customName"] = opts.CustomName
		}
		if opts.ClusterID != "" {
			data["cl"] = opts.ClusterID
		}
		if opts.Timelock != "" {
			data["timelock"] = opts.Timelock
		}
	}
	return c.do("POST", "/files/pin/"+url.PathEscape(cid), data, nil, true)
}

// PinOptions for PinCid
type PinOptions struct {
	OriginalName string
	CustomName   string
	ClusterID    string
	Timelock     string
}

// RemoveFile DELETE /files/remove/:cid
func (c *Client) RemoveFile(cid string) (*http.Response, error) {
	return c.do("DELETE", "/files/remove/"+url.PathEscape(cid), nil, nil, true)
}

// --- Users ---

// GetMe GET /users/me
func (c *Client) GetMe() (*http.Response, error) {
	return c.do("GET", "/users/me", nil, nil, true)
}

// ListUploads GET /users/me/uploads?page=&limit=
func (c *Client) ListUploads(page, limit int) (*http.Response, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	path := fmt.Sprintf("/users/me/uploads?page=%d&limit=%d", page, limit)
	return c.do("GET", path, nil, nil, true)
}

// --- Tokens (name required; label default cli-access; expiresInDays optional) ---

// GenerateToken POST /tokens/generate – name required, optional label, expiresInDays, scopes, totpCode (2FA)
func (c *Client) GenerateToken(name string, label *string, expiresInDays *int, scopes []string, totpCode *string) (*http.Response, error) {
	data := map[string]interface{}{"name": name}
	if label != nil {
		data["label"] = *label
	}
	if expiresInDays != nil {
		data["expiresInDays"] = *expiresInDays
	}
	if len(scopes) > 0 {
		data["scopes"] = scopes
	}
	if totpCode != nil && *totpCode != "" {
		data["totpCode"] = *totpCode
	}
	return c.do("POST", "/tokens/generate", data, nil, true)
}

// ListTokens GET /tokens/list
func (c *Client) ListTokens() (*http.Response, error) {
	return c.do("GET", "/tokens/list", nil, nil, true)
}

// RevokeToken DELETE /tokens/revoke/:name. Pass totpCode when account has 2FA.
func (c *Client) RevokeToken(name string, totpCode *string) (*http.Response, error) {
	var body interface{}
	if totpCode != nil && *totpCode != "" {
		body = map[string]string{"totpCode": *totpCode}
	}
	return c.do("DELETE", "/tokens/revoke/"+url.PathEscape(name), body, nil, true)
}

// --- Status ---

// GetStatus GET /status/:cid?cl=
func (c *Client) GetStatus(cid string, clusterID *string) (*http.Response, error) {
	path := "/status/" + url.PathEscape(cid)
	if clusterID != nil && *clusterID != "" {
		path += "?cl=" + url.QueryEscape(*clusterID)
	}
	return c.do("GET", path, nil, nil, true)
}

// GetAllocations GET /allocations/:cid?cl=
func (c *Client) GetAllocations(cid string, clusterID *string) (*http.Response, error) {
	path := "/allocations/" + url.PathEscape(cid)
	if clusterID != nil && *clusterID != "" {
		path += "?cl=" + url.QueryEscape(*clusterID)
	}
	return c.do("GET", path, nil, nil, true)
}

// --- Helpers ---

func (c *Client) do(method, path string, body interface{}, query url.Values, auth bool) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil && (method == "POST" || method == "PUT" || method == "DELETE") {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}
	fullPath := c.BaseURL + path
	if query != nil && len(query) > 0 {
		fullPath += "?" + query.Encode()
	}
	req, err := http.NewRequest(method, fullPath, reqBody)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers(auth) {
		for _, vv := range v {
			req.Header.Set(k, vv)
		}
	}
	if body != nil && (method == "POST" || method == "PUT" || method == "DELETE") {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp)
	}
	return resp, nil
}

func (c *Client) doMultipart(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers(true) {
		for _, vv := range v {
			req.Header.Set(k, vv)
		}
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp)
	}
	return resp, nil
}

func parseAPIError(resp *http.Response) *APIError {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	apiErr := &APIError{StatusCode: resp.StatusCode, Body: body}
	var m map[string]interface{}
	if json.Unmarshal(body, &m) == nil {
		if s, ok := m["error"].(string); ok {
			apiErr.Err = s
		}
		if s, ok := m["message"].(string); ok {
			apiErr.Message = s
		}
		if s, ok := m["code"].(string); ok {
			apiErr.Code = s
		}
		if s, ok := m["required"].(string); ok {
			apiErr.Required = s
		}
		if resp.StatusCode == 429 {
			if n, ok := m["retryAfter"].(float64); ok {
				apiErr.RetryAfterSeconds = int(n)
			}
			if apiErr.RetryAfterSeconds == 0 {
				if h := resp.Header.Get("Retry-After"); h != "" {
					var ra int
					if _, err := fmt.Sscanf(h, "%d", &ra); err == nil {
						apiErr.RetryAfterSeconds = ra
					}
				}
			}
		}
	}
	return apiErr
}
