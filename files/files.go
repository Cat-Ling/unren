// Package files provides embedded file access for UnRen templates and scripts.
// Files are embedded at compile time using Go's embed directive.
package files

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed rpy/*.rpy rpy/*.tmpl
var rpyFiles embed.FS

//go:embed python/rpatool.py python/rpatool_py2.py python/rpa.py
var rpaScripts embed.FS

//go:embed python/unrpyc_py3/*
var unrpycPy3 embed.FS

//go:embed python/unrpyc_py2/*
var unrpycPy2 embed.FS

// GetRPYContent returns the content of a static RPY file.
// For templated files, use GetRPYTemplated instead.
func GetRPYContent(filename string) ([]byte, error) {
	return rpyFiles.ReadFile("rpy/" + filename)
}

// QuickSaveConfig holds configuration for the quick save/load template.
type QuickSaveConfig struct {
	QuickSaveKey string
	QuickLoadKey string
}

// GetRPYTemplated returns the content of a templated RPY file
// after applying the provided data.
func GetRPYTemplated(filename string, data interface{}) ([]byte, error) {
	content, err := rpyFiles.ReadFile("rpy/" + filename)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(filename).Parse(string(content))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RPYFiles contains the standard RPY file names used by UnRen.
var RPYFiles = struct {
	Dev      string
	Quick    string
	Skip     string
	Rollback string
}{
	Dev:      "unren-dev.rpy",
	Quick:    "unren-quick.rpy.tmpl",
	Skip:     "unren-skip.rpy",
	Rollback: "unren-rollback.rpy",
}

// GetRPATool returns the rpatool Python script content.
// If python3 is true, returns the Python 3 version; otherwise Python 2.
func GetRPATool(python3 bool) ([]byte, error) {
	if python3 {
		return rpaScripts.ReadFile("python/rpatool.py")
	}
	return rpaScripts.ReadFile("python/rpatool_py2.py")
}

// GetRPAFallback returns the simple rpa.py fallback extractor.
func GetRPAFallback() ([]byte, error) {
	return rpaScripts.ReadFile("python/rpa.py")
}

// ExtractUnrpyc extracts the unrpyc decompiler files to the specified directory.
// If python3 is true, extracts the Python 3 version; otherwise Python 2.
func ExtractUnrpyc(destDir string, python3 bool) error {
	var fs embed.FS
	var prefix string

	if python3 {
		fs = unrpycPy3
		prefix = "python/unrpyc_py3"
	} else {
		fs = unrpycPy2
		prefix = "python/unrpyc_py2"
	}

	entries, err := fs.ReadDir(prefix)
	if err != nil {
		return err
	}

	// Create destination directory
	decompilerDir := filepath.Join(destDir, "decompiler")
	if err := os.MkdirAll(decompilerDir, 0755); err != nil {
		return err
	}

	// Extract each file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := fs.ReadFile(prefix + "/" + entry.Name())
		if err != nil {
			return err
		}

		destPath := filepath.Join(decompilerDir, entry.Name())
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return err
		}
	}

	return nil
}

// GetUnrpycMainScript returns the main unrpyc.py script content.
func GetUnrpycMainScript(python3 bool) ([]byte, error) {
	if python3 {
		return unrpycPy3.ReadFile("python/unrpyc_py3/unrpyc.py")
	}
	return unrpycPy2.ReadFile("python/unrpyc_py2/unrpyc.py")
}
