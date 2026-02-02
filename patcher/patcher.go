// Package patcher creates .rpy modification files for Ren'Py games.
// It uses embedded templates for consistency and maintainability.
package patcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/unren/unren-go/files"
)

// Config holds configuration for the patcher.
type Config struct {
	// QuickSaveKey is the Ren'Py key constant for quick save (default: K_F5)
	QuickSaveKey string
	// QuickLoadKey is the Ren'Py key constant for quick load (default: K_F9)
	QuickLoadKey string
}

// DefaultConfig returns the default patcher configuration.
func DefaultConfig() *Config {
	return &Config{
		QuickSaveKey: "K_F5",
		QuickLoadKey: "K_F9",
	}
}

// Patcher creates .rpy modification files for Ren'Py games.
type Patcher struct {
	gameDir string
	config  *Config
}

// New creates a new Patcher for the given game directory.
func New(gameDir string, config *Config) *Patcher {
	if config == nil {
		config = DefaultConfig()
	}
	return &Patcher{
		gameDir: gameDir,
		config:  config,
	}
}

// EnableConsole creates unren-dev.rpy to enable developer console and menu.
// Console: SHIFT+O | Dev Menu: SHIFT+D
func (p *Patcher) EnableConsole() error {
	content, err := files.GetRPYContent(files.RPYFiles.Dev)
	if err != nil {
		return fmt.Errorf("failed to load dev template: %w", err)
	}
	return p.writeFile("unren-dev.rpy", content)
}

// EnableQuickSave creates unren-quick.rpy to enable quick save/load hotkeys.
func (p *Patcher) EnableQuickSave() error {
	data := files.QuickSaveConfig{
		QuickSaveKey: p.config.QuickSaveKey,
		QuickLoadKey: p.config.QuickLoadKey,
	}

	content, err := files.GetRPYTemplated(files.RPYFiles.Quick, data)
	if err != nil {
		return fmt.Errorf("failed to process quick save template: %w", err)
	}

	return p.writeFile("unren-quick.rpy", content)
}

// EnableSkip creates unren-skip.rpy to enable skipping unseen content.
func (p *Patcher) EnableSkip() error {
	content, err := files.GetRPYContent(files.RPYFiles.Skip)
	if err != nil {
		return fmt.Errorf("failed to load skip template: %w", err)
	}
	return p.writeFile("unren-skip.rpy", content)
}

// EnableRollback creates unren-rollback.rpy to enable infinite rollback.
func (p *Patcher) EnableRollback() error {
	content, err := files.GetRPYContent(files.RPYFiles.Rollback)
	if err != nil {
		return fmt.Errorf("failed to load rollback template: %w", err)
	}
	return p.writeFile("unren-rollback.rpy", content)
}

// EnableAll enables all patching features (console, quick save, skip, rollback).
func (p *Patcher) EnableAll() error {
	if err := p.EnableConsole(); err != nil {
		return fmt.Errorf("console: %w", err)
	}
	if err := p.EnableQuickSave(); err != nil {
		return fmt.Errorf("quick save: %w", err)
	}
	if err := p.EnableSkip(); err != nil {
		return fmt.Errorf("skip: %w", err)
	}
	if err := p.EnableRollback(); err != nil {
		return fmt.Errorf("rollback: %w", err)
	}
	return nil
}

// PatchFiles returns the list of all patch file names.
func PatchFiles() []string {
	return []string{
		"unren-dev.rpy",
		"unren-quick.rpy",
		"unren-skip.rpy",
		"unren-rollback.rpy",
	}
}

// RemoveAll removes all unren patch files from the game directory.
func (p *Patcher) RemoveAll() error {
	for _, f := range PatchFiles() {
		path := filepath.Join(p.gameDir, f)
		if err := p.removeIfExists(path); err != nil {
			return err
		}
	}
	return nil
}

// removeIfExists removes a file if it exists, ignoring "not found" errors.
func (p *Patcher) removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", filepath.Base(path), err)
	}
	return nil
}

// writeFile writes content to a file in the game directory.
// It removes any existing file first to ensure clean state.
func (p *Patcher) writeFile(filename string, content []byte) error {
	path := filepath.Join(p.gameDir, filename)

	// Remove existing file if present
	_ = p.removeIfExists(path)

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	return nil
}
