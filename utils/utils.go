package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ClearTerminal clears the terminal screen in a cross-platform way
func ClearTerminal() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Colors for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// supportsColor checks if the terminal supports ANSI colors
func supportsColor() bool {
	// Windows cmd.exe doesn't support ANSI by default, but modern terminals do
	if runtime.GOOS == "windows" {
		// Check for Windows Terminal, VS Code, or other modern terminals
		if os.Getenv("WT_SESSION") != "" || os.Getenv("TERM_PROGRAM") != "" {
			return true
		}
		return false
	}
	// Unix-like systems generally support colors
	return os.Getenv("NO_COLOR") == ""
}

// ColorPrint prints colored text if supported
func ColorPrint(color, text string) {
	if supportsColor() {
		fmt.Print(color + text + ColorReset)
	} else {
		fmt.Print(text)
	}
}

// ColorPrintln prints colored text with newline if supported
func ColorPrintln(color, text string) {
	if supportsColor() {
		fmt.Println(color + text + ColorReset)
	} else {
		fmt.Println(text)
	}
}

// PrintError prints an error message in red
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ColorPrintln(ColorRed, "ERROR: "+msg)
}

// PrintSuccess prints a success message in green
func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ColorPrintln(ColorGreen, msg)
}

// PrintWarning prints a warning message in yellow
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ColorPrintln(ColorYellow, "WARNING: "+msg)
}

// PrintInfo prints an info message in cyan
func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ColorPrintln(ColorCyan, msg)
}

// NormalizePath normalizes a path for the current OS
func NormalizePath(path string) string {
	// Convert all separators to the current OS separator
	path = filepath.Clean(path)
	return path
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists and is not a directory
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// FindFilesWithExtension finds all files with a given extension in a directory (recursive)
func FindFilesWithExtension(dir, ext string) ([]string, error) {
	var files []string
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ext {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
