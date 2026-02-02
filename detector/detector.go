package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/unren/unren-go/utils"
)

// GameInfo contains detected information about a Ren'Py game
type GameInfo struct {
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
		return nil, &GameNotFoundError{Dir: absDir}
	}

	// Check for lib/ and renpy/ directories
	info.HasLib = utils.DirExists(filepath.Join(info.RootDir, "lib"))
	info.HasRenPy = utils.DirExists(filepath.Join(info.RootDir, "renpy"))

	// Detect Ren'Py version
	info.RenPyVersion = detectRenPyVersion(info.RootDir)

	// Find RPA files
	info.RPAFiles, _ = utils.FindFilesWithExtension(info.GameDir, ".rpa")

	// Find RPYC files
	info.RPYCFiles, _ = utils.FindFilesWithExtension(info.GameDir, ".rpyc")

	return info, nil
}

// detectRenPyVersion attempts to detect the Ren'Py version
func detectRenPyVersion(rootDir string) int {
	// Check for Python version indicators in lib/ directory
	libDir := filepath.Join(rootDir, "lib")
	if !utils.DirExists(libDir) {
		return 0
	}

	// Walk lib directory looking for python version hints
	entries, err := os.ReadDir(libDir)
	if err != nil {
		return 0
	}

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

	// Check for specific version files
	if utils.FileExists(filepath.Join(rootDir, "renpy", "__pycache__")) {
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
