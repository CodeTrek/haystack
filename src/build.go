//go:build ignore

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Target struct {
	GOOS   string
	GOARCH string
	Ext    string
}

var targets = []Target{
	{"windows", "amd64", ".exe"},
	{"windows", "arm64", ".exe"},
	{"linux", "amd64", ""},
	{"linux", "arm64", ""},
	{"darwin", "amd64", ""},
	{"darwin", "arm64", ""},
}

func main() {
	appName := "haystack"
	outputDir := "dist"
	version := getVersion()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	extPkgDir := filepath.Join(wd, "../extensions/vscode/pkgs")
	os.RemoveAll(extPkgDir)
	if err := os.MkdirAll(extPkgDir, 0755); err != nil {
		panic(err)
	}

	os.RemoveAll(outputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}

	for _, t := range targets {
		fmt.Printf("🔨 Building for %s/%s...\n", t.GOOS, t.GOARCH)

		binName := fmt.Sprintf("%s%s", appName, t.Ext)
		binPath := filepath.Join(outputDir, binName)

		ldflags := fmt.Sprintf("-s -w -X 'main.version=%s'", version)
		args := []string{
			"build",
			"-trimpath",
			"-ldflags", ldflags,
			"-gcflags=all=-l",
			"-o", binPath,
			"main.go",
		}

		cmd := exec.Command("go", args...)
		cmd.Env = append(os.Environ(),
			"GOOS="+t.GOOS,
			"GOARCH="+t.GOARCH,
			"CGO_ENABLED=0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Build failed: %v\n", err)
			continue
		}

		zipName := fmt.Sprintf("%s-%s-%s-v%s.zip", appName, t.GOOS, t.GOARCH, version)
		zipPath := filepath.Join(outputDir, zipName)

		if err := zipFile(zipPath, binPath); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Zip failed: %v\n", err)
		} else {
			fmt.Printf("✅ Built and zipped: %s\n", zipName)
		}

		_ = os.Remove(binPath)
	}

	os.WriteFile(filepath.Join(outputDir, "VERSION"), []byte(version), 0644)
}

func getVersion() string {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		fmt.Println("❌ Failed to read VERSION file:", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(data))
}

func zipFile(zipPath, filePath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	fileToZip, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}
