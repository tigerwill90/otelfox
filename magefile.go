//go:build mage
// +build mage

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	baseURL = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-go-contrib/main"
)

// DownloadSemConv downloads files from semconv directory
func DownloadSemConv() error {
	return downloadFiles("semconv")
}

// DownloadSemConvUtil downloads files from semconvutil directory
func DownloadSemConvUtil() error {
	return downloadFiles("semconvutil")
}

// DownloadAll downloads all files from both directories
func DownloadAll() error {
	for dir := range filesMap {
		if err := downloadFiles(dir); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to download files
func downloadFiles(dirKey string) error {
	dirInfo, exists := filesMap[dirKey]
	if !exists {
		return fmt.Errorf("unknown directory key: %s", dirKey)
	}

	if err := os.MkdirAll(dirInfo.DestPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	fmt.Printf("Downloading files from %s/%s\n", baseURL, dirInfo.SourcePath)

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
func ListTargetFiles() error {
	totalFiles := 0

	for dirKey, dirInfo := range filesMap {
		fmt.Printf("Files from %s to be downloaded from: %s/%s\n", dirKey, baseURL, dirInfo.SourcePath)
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
