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
			"env.go.tmpl",
			"env_test.go.tmpl",
			"httpconv.go.tmpl",
			"httpconv_test.go.tmpl",
			"util.go.tmpl",
			"util_test.go.tmpl",
			"v1.20.0.go.tmpl",
		},
		SourcePath: "internal/shared/semconv",
		DestPath:   "internal/shared/semconv",
	},
	"semconvutil": {
		Files: []string{
			"httpconv.go.tmpl",
			"httpconv_test.go.tmpl",
			"netconv.go.tmpl",
			"netconv_test.go.tmpl",
		},
		SourcePath: "internal/shared/semconvutil",
		DestPath:   "internal/shared/semconvutil",
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
		url := fmt.Sprintf("%s/%s/%s", baseURL, dirInfo.SourcePath, file)
		targetPath := filepath.Join(dirInfo.DestPath, file)

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
	}

	fmt.Printf("Successfully downloaded all files from %s\n", dirInfo.SourcePath)
	return nil
}

// ListTargetFiles lists all files that will be downloaded
// for the provided reference (tag or branch name)
func ListTargetFiles(ref string) error {
	if ref == "" {
		return fmt.Errorf("reference (tag or branch name) is required")
	}

	totalFiles := 0

	var baseURL string
	var refType string

	if isTag(ref) {
		baseURL = fmt.Sprintf(tagURLPattern, ref)
		refType = "tag"
	} else {
		baseURL = fmt.Sprintf(branchURLPattern, ref)
		refType = "branch"
	}

	sourceInfo := fmt.Sprintf("(%s: %s)", refType, ref)

	for dirKey, dirInfo := range filesMap {
		fmt.Printf("Files from %s to be downloaded from: %s/%s %s\n", dirKey, baseURL, dirInfo.SourcePath, sourceInfo)
		fmt.Printf("Destination: %s\n", dirInfo.DestPath)
		fmt.Println("=================================")

		for _, file := range dirInfo.Files {
			fmt.Println(file)
		}

		fmt.Printf("Total files in %s: %d\n", dirKey, len(dirInfo.Files))
		totalFiles += len(dirInfo.Files)
		fmt.Println("=================================")
	}

	fmt.Printf("Total files to be downloaded: %d\n", totalFiles)
	return nil
}

// isTag checks if the string is a version tag (vX.Y.Z format)
func isTag(ref string) bool {
	matched, _ := regexp.MatchString(`^v\d+\.\d+\.\d+.*$`, ref)
	return matched
}
