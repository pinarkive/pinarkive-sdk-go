# Pinarkive Go SDK

Minimal Go client for the **PinArkive API v3**. Upload files, pin by CID, manage tokens, and check status. See [pinarkive.com/docs.php](https://pinarkive.com/docs.php).

**Version:** 3.1.0

## Installation

```bash
go get github.com/pinarkive/pinarkive-sdk-go
```

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/pinarkive/pinarkive-sdk-go"
)

func main() {
	// NewClient(token, apiKey, baseURL) — empty string = default (https://api.pinarkive.com/api/v3)
	client := pinarkive.NewClient("", "your-api-key-here", "")

	// Upload a file
	resp, err := client.UploadFile("document.pdf", nil, nil)
	if err != nil {
		if apiErr, ok := err.(*pinarkive.APIError); ok {
			log.Fatalf("api [%d]: %s", apiErr.StatusCode, apiErr.Message)
		}
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Println("CID:", result["cid"])

	// List uploads
	resp, err = client.ListUploads(1, 20)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}
```

## Authentication

- **NewClient(token, apiKey, baseURL):** empty string for default base URL. API key is sent as `X-API-Key`; token as `Authorization: Bearer <token>`.
- **RequestSourceWeb:** set `client.RequestSourceWeb = true` when using the client from a web app. The SDK will send `X-Request-Source: web` on every Bearer-authenticated request (not when using API Key), so the backend classifies them as **WEB** in logs instead of **JWT**.

## API Methods (minimal set)

| Method | Description |
|--------|-------------|
| `Health()` | GET /health |
| `GetPlans()` | GET /plans/ |
| `GetPeers()` | GET /peers/ |
| `Login(email, password)` | POST /auth/login |
| `Verify2FALogin(temporaryToken, code string)` | POST /auth/2fa/verify-login |
| `UploadFile(filePath, clusterID, timelock *string)` | POST /files/ |
| `UploadDirectory(dirPath, clusterID, timelock *string)` | POST /files/directory |
| `UploadDirectoryDAG(files map[string]io.Reader, dirName string, clusterID, timelock *string)` | POST /files/directory-dag |
| `PinCid(cid string, opts *PinOptions)` | POST /files/pin/:cid |
| `RemoveFile(cid)` | DELETE /files/remove/:cid |
| `GetMe()` | GET /users/me |
| `ListUploads(page, limit int)` | GET /users/me/uploads |
| `GenerateToken(name, label, expiresInDays *int, scopes []string, totpCode *string)` | POST /tokens/generate |
| `ListTokens()` | GET /tokens/list |
| `RevokeToken(name string, totpCode *string)` | DELETE /tokens/revoke/:name |
| `GetStatus(cid string, clusterID *string)` | GET /status/:cid |
| `GetAllocations(cid string, clusterID *string)` | GET /allocations/:cid |

`PinOptions` has `OriginalName`, `CustomName`, `ClusterID`, `Timelock`. Optional pointer params can be `nil`.

## Error handling

On HTTP 4xx/5xx methods return `(nil, *APIError)`. **`APIError`** has:

- `StatusCode` — HTTP status
- `Err` — API field `error`
- `Message` — API field `message`
- `Code` — API field `code` (e.g. `email_not_verified`, `missing_scope`, `2fa_required`)
- `Required` — for 403 `missing_scope`: the required scope
- `RetryAfterSeconds` — for 429: seconds until retry (0 if not set)
- `Body` — raw response bytes

```go
resp, err := client.UploadFile("file.pdf", nil, nil)
if err != nil {
    if apiErr, ok := err.(*pinarkive.APIError); ok {
        log.Printf("status=%d message=%s code=%s", apiErr.StatusCode, apiErr.Message, apiErr.Code)
    }
    return
}
defer resp.Body.Close()
```

## Changelog

### 3.1.0

- **Request source:** `Client.RequestSourceWeb = true` sends `X-Request-Source: web` on Bearer requests (for backend logs).
- **Scopes & 2FA:** `GenerateToken` accepts `scopes []string` and `totpCode *string`; `RevokeToken` accepts `totpCode *string`. `Verify2FALogin(temporaryToken, code)` for login with 2FA.
- **Errors:** `APIError` has `Required` (403 missing_scope) and `RetryAfterSeconds` (429).

### 3.0.0

- **API v3:** Base URL is now `https://api.pinarkive.com/api/v3` (was `/api/v2`). v1/v2 are deprecated (410).
- **Errors:** On 4xx/5xx methods return `(nil, *APIError)` with `StatusCode`, `Err`, `Message`, `Code`, `Body`. Check `err != nil` and type-assert to `*pinarkive.APIError` for API details.
- **Minimal surface:** Only endpoints documented at [pinarkive.com/docs.php](https://pinarkive.com/docs.php): Health, GetPlans, GetPeers, Login, UploadFile, UploadDirectory, UploadDirectoryDAG, PinCid, RemoveFile, GetMe, ListUploads, GenerateToken, ListTokens, RevokeToken, GetStatus, GetAllocations. Optional `*string` for cluster/timelock; `PinOptions` for pin.
- **Removed:** `RenameFile`; `TokenOptions` with Permissions/IPAllowlist. Use `GenerateToken(name, label *string, expiresInDays *int)`.
- **Client:** Constructor is `NewClient(token, apiKey, baseURL)`; empty baseURL defaults to v3.
- **Pin:** `PinCid(cid, opts *PinOptions)` with `OriginalName`, `CustomName`, `ClusterID`, `Timelock` (replacing the old `PinCid(cid, filename)`).
- **Uploads:** `UploadFile` and `UploadDirectory` take optional `clusterID, timelock *string`; `UploadDirectoryDAG` sends content as file parts.

### Upgrading from 2.x

1. Switch to `NewClient(token, apiKey, baseURL)` and use default `""` for v3 base URL.
2. Check errors: on 4xx/5xx you get `(nil, *APIError)`; handle with `if apiErr, ok := err.(*pinarkive.APIError)`.
3. Use `PinCid(cid, &pinarkive.PinOptions{CustomName: "x", ClusterID: "cl0"})` instead of `PinCid(cid, filename)`.
4. Use `GenerateToken(name, &label, &expiresInDays)` instead of `GenerateToken(name, TokenOptions{...})`.
5. Use `UploadFile(path, clusterID, timelock)` and `UploadDirectory(path, clusterID, timelock)` with pointers; pass `nil` when not used.
6. Pin with `go get github.com/pinarkive/pinarkive-sdk-go@v3.1.0` (or later).

## Links

- [API docs](https://pinarkive.com/docs.php)
- [Repository](https://github.com/pinarkive/pinarkive-sdk-go)
