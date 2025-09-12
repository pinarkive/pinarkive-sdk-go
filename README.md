# Pinarkive Go SDK

Go client for the Pinarkive API v2.3.1. Easy IPFS file management with directory DAG uploads, file renaming, and enhanced API key management. Perfect for Go web applications and microservices.

## Installation

### Using Go Modules (Recommended)

```bash
go get github.com/pinarkive/pinarkive-sdk-go
```

### Manual Installation

Copy `pinarkive_client.go` to your project and ensure you have Go 1.19+.

## Quick Start

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/pinarkive/pinarkive-sdk-go"
)

func main() {
	// Initialize with API key
	client := pinarkive.NewPinarkiveClient("", "your-api-key-here", "")

	// Upload a file
	resp, err := client.UploadFile("document.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("File uploaded: %s\n", string(body))

	// Generate API key
	tokenResp, err := client.GenerateToken("my-app", pinarkive.TokenOptions{
		ExpiresInDays: 30,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tokenResp.Body.Close()

	tokenBody, _ := ioutil.ReadAll(tokenResp.Body)
	fmt.Printf("New API key: %s\n", string(tokenBody))
}
```

## Authentication

The SDK supports two authentication methods:

### API Key Authentication (Recommended)
```go
client := pinarkive.NewPinarkiveClient("", "your-api-key-here", "")
```
**Note:** The SDK automatically sends the API key using the `Authorization: Bearer` header format, not `X-API-Key`.

### JWT Token Authentication
```go
client := pinarkive.NewPinarkiveClient("your-jwt-token-here", "", "")
```

## Basic Usage

### File Upload
```go
// Upload single file
resp, err := client.UploadFile("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("CID: %s\n", string(body))
```

### Directory Upload
```go
// Upload directory from local path
resp, err := client.UploadDirectory("/path/to/directory")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("Directory CID: %s\n", string(body))
```

### List Uploads
```go
// List all uploaded files with pagination
resp, err := client.ListUploads(1, 20)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("Uploads: %s\n", string(body))
```

## Advanced Features

### Directory DAG Upload
Upload entire directory structures as DAG (Directed Acyclic Graph):

```go
// Create project structure
files := map[string]io.Reader{
    "src/main.go":    strings.NewReader("package main\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}"),
    "src/utils.go":   strings.NewReader("package main\n\nfunc utils() {}\n"),
    "go.mod":         strings.NewReader("module my-project\n\ngo 1.19\n"),
    "README.md":      strings.NewReader("# My Project\n\nThis is my project."),
}

// Upload as DAG
resp, err := client.UploadDirectoryDAG(files, "my-project")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("DAG CID: %s\n", string(body))
```

### Directory Cluster Upload
```go
// Upload using cluster-based approach
files := []pinarkive.FileUpload{
    {
        Path:    "file1.txt",
        Content: strings.NewReader("Content 1"),
    },
    {
        Path:    "file2.txt",
        Content: strings.NewReader("Content 2"),
    },
}

resp, err := client.UploadDirectoryCluster(files)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("Cluster CID: %s\n", string(body))
```

### Upload File to Existing Directory
```go
// Add file to existing directory
resp, err := client.UploadFileToDirectory("new-file.txt", "existing-directory-path")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("File added to directory: %s\n", string(body))
```

### File Renaming
```go
// Rename an uploaded file
resp, err := client.RenameFile("upload-id-here", "new-file-name.pdf")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("File renamed: %s\n", string(body))
```

### File Removal
```go
// Remove a file from storage
resp, err := client.RemoveFile("QmYourCIDHere")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("File removed: %s\n", string(body))
```

### Pinning Operations

#### Basic CID Pinning
```go
// Pin with filename
resp, err := client.PinCid("QmYourCIDHere", "my-file.pdf")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("CID pinned: %s\n", string(body))

// Pin without filename (backend will use default)
resp2, err := client.PinCid("QmYourCIDHere", "")
if err != nil {
    log.Fatal(err)
}
defer resp2.Body.Close()

body2, _ := ioutil.ReadAll(resp2.Body)
fmt.Printf("CID pinned: %s\n", string(body2))
```

#### Pin with Custom Name
```go
resp, err := client.PinCidWithName("QmYourCIDHere", "my-important-file")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("CID pinned with name: %s\n", string(body))
```

### API Key Management

#### Generate API Key
```go
// Basic token generation
resp, err := client.GenerateToken("my-app", pinarkive.TokenOptions{})
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

// Advanced token with options
resp, err = client.GenerateToken("my-app", pinarkive.TokenOptions{
    ExpiresInDays: 30,
    IPAllowlist:   []string{"192.168.1.1", "10.0.0.1"},
    Permissions:   []string{"upload", "pin"},
})
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("New API key: %s\n", string(body))
```

#### List API Keys
```go
resp, err := client.ListTokens()
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("API Keys: %s\n", string(body))
```

#### Revoke API Key
```go
resp, err := client.RevokeToken("my-app")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("Token revoked: %s\n", string(body))
```

## Error Handling

```go
resp, err := client.UploadFile("document.pdf")
if err != nil {
    log.Printf("Network error: %v", err)
    return
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    body, _ := ioutil.ReadAll(resp.Body)
    log.Printf("API Error: %d - %s", resp.StatusCode, string(body))
    return
}

body, _ := ioutil.ReadAll(resp.Body)
fmt.Printf("Success: %s\n", string(body))
```

## Framework Integration

### Gin Web Framework
```go
package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pinarkive/pinarkive-sdk-go"
)

func main() {
	r := gin.Default()
	client := pinarkive.NewPinarkiveClient("", os.Getenv("PINARKIVE_API_KEY"), "")

	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
			return
		}

		// Save uploaded file temporarily
		tempPath := "/tmp/" + file.Filename
		if err := c.SaveUploadedFile(file, tempPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer os.Remove(tempPath)

		// Upload to Pinarkive
		resp, err := client.UploadFile(tempPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		c.Data(http.StatusOK, "application/json", body)
	})

	r.GET("/files", func(c *gin.Context) {
		resp, err := client.ListUploads(1, 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		c.Data(http.StatusOK, "application/json", body)
	})

	r.Run(":8080")
}
```

### Echo Web Framework
```go
package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/pinarkive/pinarkive-sdk-go"
)

func main() {
	e := echo.New()
	client := pinarkive.NewPinarkiveClient("", os.Getenv("PINARKIVE_API_KEY"), "")

	e.POST("/upload", func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "No file provided"})
		}

		// Save uploaded file temporarily
		tempPath := "/tmp/" + file.Filename
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer src.Close()

		dst, err := os.Create(tempPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer dst.Close()
		defer os.Remove(tempPath)

		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Upload to Pinarkive
		resp, err := client.UploadFile(tempPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		return c.Blob(http.StatusOK, "application/json", body)
	})

	e.GET("/files", func(c echo.Context) error {
		resp, err := client.ListUploads(1, 10)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		return c.Blob(http.StatusOK, "application/json", body)
	})

	e.Start(":8080")
}
```

## API Reference

### Constructor
```go
func NewPinarkiveClient(token, apiKey, baseURL string) *PinarkiveClient
```
- `token`: JWT token for authentication (can be empty)
- `apiKey`: API key for authentication (can be empty)
- `baseURL`: Base URL for the API (can be empty for default)

### File Operations
- `UploadFile(filePath string)` - Upload single file
- `UploadDirectory(dirPath string)` - Upload directory recursively (calls UploadFile for each file)
- `UploadDirectoryDAG(files map[string]io.Reader, dirName string)` - Upload directory as DAG structure
- `RenameFile(uploadID, newName string)` - Rename uploaded file
- `RemoveFile(cid string)` - Remove file from storage

### Pinning Operations
- `PinCid(cid, filename string)` - Pin CID to account with filename

### User Operations
- `ListUploads(page, limit int)` - List uploaded files

### Token Management
- `GenerateToken(name string, options TokenOptions)` - Generate API key
- `ListTokens()` - List all API keys
- `RevokeToken(name string)` - Revoke API key


### Status & Monitoring
- `GetStatus(cid string)` - Get file status
- `GetAllocations(cid string)` - Get storage allocations

## Examples

### Complete File Management Workflow
```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/pinarkive/pinarkive-sdk-go"
)

func main() {
	client := pinarkive.NewPinarkiveClient("", "your-api-key", "")

	// 1. Upload a file
	resp, err := client.UploadFile("document.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var uploadData map[string]interface{}
	json.Unmarshal(body, &uploadData)
	fmt.Printf("Uploaded: %v\n", uploadData["cid"])

	// 2. Pin the CID with a custom name
	pinResp, err := client.PinCidWithName(uploadData["cid"].(string), "important-document")
	if err != nil {
		log.Fatal(err)
	}
	defer pinResp.Body.Close()

	pinBody, _ := ioutil.ReadAll(pinResp.Body)
	var pinData map[string]interface{}
	json.Unmarshal(pinBody, &pinData)
	fmt.Printf("Pinned: %v\n", pinData["pinned"])

	// 3. Rename the file
	if uploadID, ok := uploadData["uploadId"].(string); ok {
		renameResp, err := client.RenameFile(uploadID, "my-document.pdf")
		if err != nil {
			log.Fatal(err)
		}
		defer renameResp.Body.Close()

		renameBody, _ := ioutil.ReadAll(renameResp.Body)
		var renameData map[string]interface{}
		json.Unmarshal(renameBody, &renameData)
		fmt.Printf("Renamed: %v\n", renameData["updated"])
	}

	// 4. List all uploads
	uploadsResp, err := client.ListUploads(1, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer uploadsResp.Body.Close()

	uploadsBody, _ := ioutil.ReadAll(uploadsResp.Body)
	var uploadsData map[string]interface{}
	json.Unmarshal(uploadsBody, &uploadsData)
	fmt.Printf("All uploads: %v\n", uploadsData["uploads"])
}
```

### Directory Upload Workflow
```go
package main

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/pinarkive/pinarkive-sdk-go"
)

func main() {
	client := pinarkive.NewPinarkiveClient("", "your-api-key", "")

	// Create project structure
	files := map[string]io.Reader{
		"src/main.go":    strings.NewReader("package main\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}"),
		"src/utils.go":   strings.NewReader("package main\n\nfunc utils() {}\n"),
		"go.mod":         strings.NewReader("module my-project\n\ngo 1.19\n"),
		"README.md":      strings.NewReader("# My Project\n\nThis is my project."),
	}

	resp, err := client.UploadDirectoryDAG(files, "my-project")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Project uploaded: %s\n", string(body))
}
```

### Concurrent File Processing
```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pinarkive/pinarkive-sdk-go"
)

func main() {
	client := pinarkive.NewPinarkiveClient("", "your-api-key", "")

	// Process multiple files concurrently
	filePaths := []string{"file1.txt", "file2.txt", "file3.txt"}
	var wg sync.WaitGroup
	results := make(chan string, len(filePaths))

	for _, filePath := range filePaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			resp, err := client.UploadFile(path)
			if err != nil {
				log.Printf("Error uploading %s: %v", path, err)
				return
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			results <- fmt.Sprintf("%s: %s", path, string(body))
		}(filePath)
	}

	wg.Wait()
	close(results)

	for result := range results {
		fmt.Println(result)
	}
}
```

## Publishing Instructions

### Publishing Go Modules

Go modules are automatically published when you push to the main branch:

```bash
# Update version in go.mod if needed (usually not required for Go modules)
git add go.mod
git commit -m "Update module path for v2.3.1"
git tag v2.3.1
git push origin main --tags
# Users can then: go get github.com/pinarkive/pinarkive-sdk-go@v2.3.1
```

### Version Management

- Create git tag with format `v2.3.1`
- Push to main branch with tags
- Go modules automatically use semantic versioning
- Users can install specific versions: `go get github.com/pinarkive/pinarkive-sdk-go@v2.3.1`

### Go Module Best Practices

- Use semantic versioning (v1.0.0, v2.0.0, etc.)
- Tag releases with `v` prefix
- Keep go.mod minimal with only necessary dependencies
- Document breaking changes in release notes

## Support

For issues or questions:
- GitHub Issues: [https://github.com/pinarkive/pinarkive-sdk-go/issues](https://github.com/pinarkive/pinarkive-sdk-go/issues)
- API Documentation: [https://api.pinarkive.com/docs](https://api.pinarkive.com/docs)
- Contact: [https://pinarkive.com/docs.php](https://pinarkive.com/docs.php) 