//go:build mage
// +build mage

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// Whitelisted files for download
var filesMap = map[string]struct {
	Files      []string
	SourcePath string
	DestPath   string
}{
	"semconv": {
		Files: []string{
			"bench_test.go.tmpl",
			"client.go.tmpl",
			"client_test.go.tmpl",
			"common_test.go.tmpl",
			//"env.go.tmpl",
			//"httpconv.go.tmpl",
			"httpconvtest_test.go.tmpl",
			"server.go.tmpl",
			"server_test.go.tmpl",
			"util.go.tmpl",
			"util_test.go.tmpl",
		},
		SourcePath: "internal/shared/semconv",
		DestPath:   "internal/shared/semconv",
	},
}

const (
	branchURLPattern = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-go-contrib/refs/heads/%s"
	tagURLPattern    = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-go-contrib/refs/tags/%s"
)

// DownloadSemConv downloads files from semconv directory
// using the provided reference (tag or branch name)
func DownloadSemConv(ref string) error {
	if ref == "" {
		return fmt.Errorf("reference (tag or branch name) is required")
	}
	return downloadFiles("semconv", ref)
}

// DownloadSemConvUtil downloads files from semconvutil directory
// using the provided reference (tag or branch name)
func DownloadSemConvUtil(ref string) error {
	if ref == "" {
		return fmt.Errorf("reference (tag or branch name) is required")
	}
	return downloadFiles("semconvutil", ref)
}

// DownloadAll downloads all files from both directories
// using the provided reference (tag or branch name)
func DownloadAll(ref string) error {
	if ref == "" {
		return fmt.Errorf("reference (tag or branch name) is required")
	}
	for dir := range filesMap {
		if err := downloadFiles(dir, ref); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to download files
func downloadFiles(dirKey string, ref string) error {
	dirInfo, exists := filesMap[dirKey]
	if !exists {
		return fmt.Errorf("unknown directory key: %s", dirKey)
	}

	if err := os.RemoveAll(dirInfo.DestPath); err != nil {
		fmt.Printf("Failed to delete target directory: %s", err)
	}

	if err := os.MkdirAll(dirInfo.DestPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	var baseURL string
	var refType string

	if isTag(ref) {
		baseURL = fmt.Sprintf(tagURLPattern, ref)
		refType = "tag"
	} else {
		baseURL = fmt.Sprintf(branchURLPattern, ref)
		refType = "branch"
	}

	fmt.Printf("Downloading files from %s/%s (%s: %s)\n", baseURL, dirInfo.SourcePath, refType, ref)

	for _, file := range dirInfo.Files {
		if err := downloadFile(baseURL, dirInfo.SourcePath, dirInfo.DestPath, file); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully downloaded all files from %s\n", dirInfo.SourcePath)
	return nil
}

func downloadFile(baseURL, sourcePath, destPath, file string) error {
	url := fmt.Sprintf("%s/%s/%s", baseURL, sourcePath, file)
	targetPath := filepath.Join(destPath, file)

	fmt.Printf("Downloading %s to %s\n", url, targetPath)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", file, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response code: %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", targetPath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", targetPath, err)
	}

	return nil
}

// isTag checks if the string is a version tag (vX.Y.Z format)
func isTag(ref string) bool {
	matched, _ := regexp.MatchString(`^v\d+\.\d+\.\d+.*$`, ref)
	return matched
}
