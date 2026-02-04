package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unren/unren-go/utils"
)

// GameInfo contains detected information about a Ren'Py game
type GameInfo struct {
	// Name of the game (for display purposes)
	Name string
	// RootDir is the game's root directory (where the executable is)
	RootDir string
	// GameDir is the game's "game" subdirectory
	GameDir string
	// RenPyVersion detected Ren'Py major version (7 or 8, 0 if unknown)
	RenPyVersion int
	// RPAFiles found in the game directory
	RPAFiles []string
	// RPYCFiles found in the game directory
	RPYCFiles []string
	// HasLib indicates if lib/ directory exists
	HasLib bool
	// LibDir is the path to the lib/ directory
	LibDir string
	// HasRenPy indicates if renpy/ directory exists
	HasRenPy bool
}

// DetectGame attempts to detect a Ren'Py game from the given directory
// It can be called from either the game root or the game/ subdirectory
func DetectGame(dir string) (*GameInfo, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	info := &GameInfo{}

	// Check for macOS App Bundle structure
	// Standard Ren'Py Mac apps: MyApp.app/Contents/Resources/autorun/game
	macAutorun := filepath.Join(absDir, "Contents", "Resources", "autorun")
	if utils.DirExists(macAutorun) {
		// Capture the .app name before switching context
		info.Name = filepath.Base(absDir)
		// Switch context to the autorun folder which acts as the game root
		absDir = macAutorun
	}

	// Check if we're in the game/ subdirectory or the root
	if filepath.Base(absDir) == "game" {
		// We're in the game/ directory, parent is root
		info.GameDir = absDir
		info.RootDir = filepath.Dir(absDir)
	} else if utils.DirExists(filepath.Join(absDir, "game")) {
		// We're in the root directory
		info.RootDir = absDir
		info.GameDir = filepath.Join(absDir, "game")
	} else {
		// Try to find game/ in current directory
		// Maybe we're in some other subdirectory
		fmt.Printf("Detection: Failed to find game directory in %s\n", absDir)
		return nil, &GameNotFoundError{Dir: absDir}
	}

	// If Name wasn't set by Mac detection, use the internal root dir name
	if info.Name == "" {
		info.Name = filepath.Base(info.RootDir)
	}

	// Check for lib/ and renpy/ directories
	libDir := filepath.Join(info.RootDir, "lib")
	if utils.DirExists(libDir) {
		info.HasLib = true
		info.LibDir = libDir
	} else {
		// Fallback for macOS: lib/ might be in the parent directory (Contents/Resources/lib)
		// while root is Contents/Resources/autorun
		parentLib := filepath.Join(filepath.Dir(info.RootDir), "lib")
		if utils.DirExists(parentLib) {
			info.HasLib = true
			info.LibDir = parentLib
		}
	}

	info.HasRenPy = utils.DirExists(filepath.Join(info.RootDir, "renpy"))

	// Detect Ren'Py version
	info.RenPyVersion = detectRenPyVersion(info)

	// Find RPA files
	info.RPAFiles, _ = utils.FindFilesWithExtension(info.GameDir, ".rpa")

	// Find RPYC files
	info.RPYCFiles, _ = utils.FindFilesWithExtension(info.GameDir, ".rpyc")

	return info, nil
}

// detectRenPyVersion attempts to detect the Ren'Py version
func detectRenPyVersion(info *GameInfo) int {
	// Check for Python version indicators in lib/ directory
	if info.HasLib {
		// Walk lib directory looking for python version hints
		entries, err := os.ReadDir(info.LibDir)
		if err == nil {
			for _, entry := range entries {
				name := strings.ToLower(entry.Name())
				// Ren'Py 8 uses Python 3
				if strings.Contains(name, "py3") || strings.Contains(name, "python3") {
					return 8
				}
				// Ren'Py 7 and earlier use Python 2
				if strings.Contains(name, "py2") || strings.Contains(name, "python2") {
					return 7
				}
			}
		}
	}

	// Check for specific version files
	if utils.FileExists(filepath.Join(info.RootDir, "renpy", "__pycache__")) {
		// __pycache__ indicates Python 3, so Ren'Py 8
		return 8
	}

	return 0
}

// HasRPAFiles returns true if the game has RPA archive files
func (g *GameInfo) HasRPAFiles() bool {
	return len(g.RPAFiles) > 0
}

// HasRPYCFiles returns true if the game has compiled RPYC files
func (g *GameInfo) HasRPYCFiles() bool {
	return len(g.RPYCFiles) > 0
}

// GameNotFoundError is returned when a valid Ren'Py game cannot be detected
type GameNotFoundError struct {
	Dir string
}

func (e *GameNotFoundError) Error() string {
	return "could not detect Ren'Py game in directory: " + e.Dir
}
