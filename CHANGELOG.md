# Changelog

All notable changes to `pinarkive-sdk-go` are documented here.

## [3.1.2] - 2026-04-14

### Changed

- **Go modules:** Module path is now `github.com/pinarkive/pinarkive-sdk-go/v3` so tags **`v3.x.x`** resolve correctly on [pkg.go.dev](https://pkg.go.dev/github.com/pinarkive/pinarkive-sdk-go/v3). Import with `import "github.com/pinarkive/pinarkive-sdk-go/v3"`.
- **Documentation:** Links updated to [https://docs.pinarkive.com](https://docs.pinarkive.com).

## [3.1.1] - 2026-04-14

### Fixed

- **`UploadDirectoryDAG` multipart format:** The API expects multer field **`files`** (repeated), with each part’s **filename** set to the relative path inside the DAG. The SDK previously sent `files[i][path]` / `files[i][content]`, which the backend does not parse into `req.files`.

### Release / publish

This repo currently has a `.github/workflows/build.yml` that builds on `main` pushes and `v*` tags, but does **not** publish to a package registry. To publish, add the appropriate workflow (GitHub Releases, Go module proxy tags, etc.) per the project’s desired distribution method.

