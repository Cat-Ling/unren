// Package runner provides functionality to execute embedded Python scripts
// using the game's bundled Python interpreter, similar to the original UnRen batch.
package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/unren/unren-go/detector"
	"github.com/unren/unren-go/files"
)

// Runner manages extraction and execution of embedded Python scripts.
type Runner struct {
	GameInfo  *detector.GameInfo
	TempDir   string
	PythonExe string
	PythonLib string
	IsPython3 bool
}

// NewRunner creates a new runner for the detected game.
func NewRunner(gameInfo *detector.GameInfo) (*Runner, error) {
	r := &Runner{
		GameInfo: gameInfo,
	}

	// Find Python executable
	pythonExe, pythonLib, isPy3, err := r.findPython()
	if err != nil {
		return nil, fmt.Errorf("failed to find Python: %w", err)
	}
	r.PythonExe = pythonExe
	r.PythonLib = pythonLib
	r.IsPython3 = isPy3

	return r, nil
}

// findPython locates the Python interpreter bundled with the game.
// Returns: pythonExe path, pythonLib path, isPython3, error
func (r *Runner) findPython() (string, string, bool, error) {
	if r.GameInfo.RootDir == "" {
		return "", "", false, fmt.Errorf("game root directory not set")
	}

	libDir := filepath.Join(r.GameInfo.RootDir, "lib")
	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		return "", "", false, fmt.Errorf("lib directory not found at %s", libDir)
	}

	// Determine OS prefix and bitness
	var osPrefix string
	switch runtime.GOOS {
	case "windows":
		osPrefix = "windows"
	case "linux":
		osPrefix = "linux"
	case "darwin":
		osPrefix = "mac"
	default:
		osPrefix = "linux"
	}

	bitness := "x86_64"
	if runtime.GOARCH != "amd64" {
		bitness = "i686"
	}

	// Try Python 3 paths first (Ren'Py 8)
	py3Patterns := []string{
		filepath.Join(libDir, fmt.Sprintf("py3-%s-%s", osPrefix, bitness)),
		filepath.Join(libDir, fmt.Sprintf("python3-%s-%s", osPrefix, bitness)),
	}

	for _, pyDir := range py3Patterns {
		pythonExe := r.getPythonExe(pyDir)
		if pythonExe != "" {
			pythonLib := r.findPythonLib(pyDir)
			return pythonExe, pythonLib, true, nil
		}
	}

	// Try Python 2 paths (Ren'Py 7)
	py2Patterns := []string{
		filepath.Join(libDir, fmt.Sprintf("py2-%s-%s", osPrefix, bitness)),
		filepath.Join(libDir, fmt.Sprintf("python2-%s-%s", osPrefix, bitness)),
		filepath.Join(libDir, fmt.Sprintf("%s-%s", osPrefix, bitness)), // Legacy
	}

	for _, pyDir := range py2Patterns {
		pythonExe := r.getPythonExe(pyDir)
		if pythonExe != "" {
			pythonLib := r.findPythonLib(pyDir)
			return pythonExe, pythonLib, false, nil
		}
	}

	// Fallback: search recursively
	pythonExe, pythonLib, isPy3 := r.searchForPython(libDir)
	if pythonExe != "" {
		return pythonExe, pythonLib, isPy3, nil
	}

	return "", "", false, fmt.Errorf("could not find Python executable in %s", libDir)
}

// getPythonExe returns the python executable path if it exists.
func (r *Runner) getPythonExe(pyDir string) string {
	var exeName string
	if runtime.GOOS == "windows" {
		exeName = "python.exe"
	} else {
		exeName = "python"
	}

	pythonPath := filepath.Join(pyDir, exeName)
	if _, err := os.Stat(pythonPath); err == nil {
		return pythonPath
	}
	return ""
}

// searchForPython recursively searches for python executable.
func (r *Runner) searchForPython(libDir string) (string, string, bool) {
	pythonNames := []string{"python", "python.exe", "python3", "python3.exe"}

	var found string
	var foundDir string
	var isPy3 bool

	filepath.Walk(libDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		for _, name := range pythonNames {
			if strings.EqualFold(info.Name(), name) {
				found = path
				foundDir = filepath.Dir(path)
				isPy3 = strings.Contains(strings.ToLower(path), "py3") ||
					strings.Contains(strings.ToLower(path), "python3")
				return filepath.SkipAll
			}
		}
		return nil
	})

	if found != "" {
		pythonLib := r.findPythonLib(foundDir)
		return found, pythonLib, isPy3
	}
	return "", "", false
}

// findPythonLib finds the Python library directory containing encodings/.
func (r *Runner) findPythonLib(pythonHome string) string {
	// Look for pythonXX or python3.X directories with encodings
	entries, err := os.ReadDir(pythonHome)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(strings.ToLower(entry.Name()), "python") {
			libPath := filepath.Join(pythonHome, entry.Name())
			if _, err := os.Stat(filepath.Join(libPath, "encodings")); err == nil {
				return libPath
			}
		}
	}

	// Also check lib/pythonX.Y paths
	var foundLib string
	filepath.Walk(pythonHome, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if strings.HasPrefix(strings.ToLower(info.Name()), "python") {
			if _, err := os.Stat(filepath.Join(path, "encodings")); err == nil {
				foundLib = path
				return filepath.SkipAll
			}
		}
		return nil
	})

	return foundLib
}

// SetupTempDir creates a temporary directory and extracts scripts.
// SetupTempDir creates a temporary directory and extracts scripts.
// The batch script uses the game root directory (maindir) for these scripts.
func (r *Runner) SetupTempDir() error {
	// Use game root directory
	r.TempDir = r.GameInfo.RootDir

	// Extract rpatool to _rpatool.py (matching batch script)
	rpatool, err := files.GetRPATool(r.IsPython3)
	if err != nil {
		return fmt.Errorf("failed to get rpatool: %w", err)
	}
	if err := os.WriteFile(filepath.Join(r.TempDir, "_rpatool.py"), rpatool, 0644); err != nil {
		return fmt.Errorf("failed to write rpatool: %w", err)
	}

	// Extract rpa.py fallback to _rpa.py (matching batch script)
	rpaFallback, err := files.GetRPAFallback()
	if err != nil {
		return fmt.Errorf("failed to get rpa.py: %w", err)
	}
	if err := os.WriteFile(filepath.Join(r.TempDir, "_rpa.py"), rpaFallback, 0644); err != nil {
		return fmt.Errorf("failed to write rpa.py: %w", err)
	}

	return nil
}

// SetupUnrpyc extracts the unrpyc decompiler.
// SetupUnrpyc extracts the unrpyc decompiler.
func (r *Runner) SetupUnrpyc() error {
	if r.TempDir == "" {
		if err := r.SetupTempDir(); err != nil {
			return err
		}
	}

	return files.ExtractUnrpyc(r.TempDir, r.IsPython3)
}

// Cleanup removes the temporary directory.
// Cleanup removes the temporary files.
func (r *Runner) Cleanup() {
	if r.TempDir == "" {
		return
	}

	files := []string{
		"_rpatool.py",
		"_rpa.py",
		"unrpyc.py",
		"deobfuscate.py",
		"_rpatool.pyc", // Python compilation files
		"_rpa.pyc",
		"unrpyc.pyc",
		"deobfuscate.pyc",
	}

	for _, f := range files {
		os.Remove(filepath.Join(r.TempDir, f))
	}

	os.RemoveAll(filepath.Join(r.TempDir, "decompiler"))
	// Also remove __pycache__ if created
	os.RemoveAll(filepath.Join(r.TempDir, "__pycache__"))
}

// getPythonEnv returns the environment variables for Python execution.
// getPythonEnv returns the environment variables for Python execution.
func (r *Runner) getPythonEnv() []string {
	pythonHome := filepath.Dir(r.PythonExe)

	env := os.Environ()
	env = append(env, fmt.Sprintf("PYTHONHOME=%s", pythonHome))

	// Build PYTHONPATH
	// Batch script sets: PYTHONPATH=%pythondir%;%pythonlibdir%;%maindir%;%decompilerdir%\
	paths := []string{pythonHome}
	if r.PythonLib != "" {
		paths = append(paths, r.PythonLib)
	}
	// Add game root (maindir) and decompiler dir
	if r.TempDir != "" {
		paths = append(paths, r.TempDir)
		paths = append(paths, filepath.Join(r.TempDir, "decompiler"))
	} else {
		// Fallback if TempDir not set (shouldn't happen if Setup called)
		paths = append(paths, r.GameInfo.RootDir)
		paths = append(paths, filepath.Join(r.GameInfo.RootDir, "decompiler"))
	}

	sep := string(os.PathListSeparator)
	env = append(env, fmt.Sprintf("PYTHONPATH=%s", strings.Join(paths, sep)))

	return env
}

// ExtractRPA extracts an RPA archive using the embedded rpatool.
func (r *Runner) ExtractRPA(rpaPath string) error {
	if r.TempDir == "" {
		if err := r.SetupTempDir(); err != nil {
			return err
		}
	}

	rpatoolPath := filepath.Join(r.TempDir, "_rpatool.py")
	rpaFallbackPath := filepath.Join(r.TempDir, "_rpa.py")
	outputDir := filepath.Dir(rpaPath)
	env := r.getPythonEnv()

	// First try rpatool.py with extract flag
	cmd := exec.Command(r.PythonExe, "-O", rpatoolPath, "-x", rpaPath, "-o", outputDir)
	cmd.Dir = outputDir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Fallback to rpa.py (simpler extractor)
		fmt.Println("    Retrying with fallback extractor...")
		cmd = exec.Command(r.PythonExe, rpaFallbackPath, rpaPath)
		cmd.Dir = outputDir
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract RPA: %w", err)
		}
	}

	return nil
}

// DecompileRPYC decompiles an RPYC file to RPY.
func (r *Runner) DecompileRPYC(rpycPath string) error {
	if err := r.SetupUnrpyc(); err != nil {
		return fmt.Errorf("failed to setup unrpyc: %w", err)
	}

	unrpycPath := filepath.Join(r.TempDir, "unrpyc.py")
	env := r.getPythonEnv()

	args := []string{"-O", unrpycPath}
	if !r.IsPython3 {
		args = append(args, "--init-offset")
	}
	args = append(args, rpycPath)

	cmd := exec.Command(r.PythonExe, args...)
	cmd.Dir = filepath.Dir(rpycPath)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// DecompileAllRPYC decompiles all RPYC files in the game directory.
// Returns: success count, skipped count, failed count, error
func (r *Runner) DecompileAllRPYC() (int, int, int, error) {
	if err := r.SetupUnrpyc(); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to setup unrpyc: %w", err)
	}

	unrpycPath := filepath.Join(r.TempDir, "unrpyc.py")
	env := r.getPythonEnv()
	gameDir := r.GameInfo.GameDir

	success := 0
	skipped := 0
	failed := 0
	lastDir := ""

	err := filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".rpyc") {
			return nil
		}

		// Skip un.rpyc (special Ren'Py file)
		baseName := filepath.Base(path)
		nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		if strings.ToLower(nameWithoutExt) == "un" {
			return nil
		}

		// Show directory change (like batch script)
		currentDir := filepath.Dir(path)
		if currentDir != lastDir {
			lastDir = currentDir
			relDir, _ := filepath.Rel(r.GameInfo.RootDir, currentDir)
			if relDir == "" {
				relDir = "."
			}
			fmt.Printf("  Working in: '%s'\n", relDir)
		}

		// Check if rpy already exists
		rpyPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".rpy"
		if _, err := os.Stat(rpyPath); err == nil {
			// Match batch script output: "filename.rpy" already exists - skipped
			fmt.Printf("    \"%s.rpy\" already exists - skipped\n", nameWithoutExt)
			skipped++
			return nil
		}

		// Show decompiling message with file size (like batch script)
		fmt.Printf("    + Decompiling \"%s\" - %d bytes\n", baseName, info.Size())

		args := []string{"-O", unrpycPath}
		if !r.IsPython3 {
			args = append(args, "--init-offset")
		}
		args = append(args, path)

		cmd := exec.Command(r.PythonExe, args...)
		cmd.Dir = filepath.Dir(path)
		cmd.Env = env

		if output, err := cmd.CombinedOutput(); err == nil {
			// Verify output file was created
			if _, err := os.Stat(rpyPath); err == nil {
				success++
			} else {
				fmt.Printf("    - Failed to create RPY file: %s.rpy not found.\n", nameWithoutExt)
				if len(output) > 0 {
					fmt.Printf("    Output:\n%s\n", string(output))
				}
				failed++
			}
		} else {
			fmt.Printf("    - Failed to decompile \"%s\". Error: %v\n", baseName, err)
			if len(output) > 0 {
				fmt.Printf("    Output:\n%s\n", string(output))
			}
			failed++
		}

		return nil
	})

	return success, skipped, failed, err
}
