# Gems & Hacks

This document outlines the specific workarounds and "hacky" solutions implemented in `unren-go` to support complex features like cross-platform macOS detection and robust automation piping.

## 1. macOS `.app` Transparency (The Detector Hack)
**Using macOS Game Files on Linux/Windows**

Ren'Py games on macOS are packaged as `.app` bundles, which are directories. The actual game assets (RPYC/RPA) live deep inside at `Contents/Resources/autorun/`.
To make this transparent to the user (so they can just point at `Game.app`), we implemented a path redirection hack in `detector/detector.go`.

**Location:** `detector/detector.go` -> `DetectGame`
**The Hack:**
```go
// Check for macOS App Bundle structure
if strings.HasSuffix(absDir, ".app") || strings.Contains(absDir, ".app/") {
    // ...
    macAutorun := filepath.Join(absDir, "Contents", "Resources", "autorun")
    if utils.DirExists(macAutorun) {
        // Switch context to the autorun folder which acts as the game root
        absDir = macAutorun
    }
}
```
We aggressively rewrite the `absDir` variable if we suspect a macOS bundle. This allows the rest of the detection logic (checking for `game/`) to work unchanged.

## 2. OS-Agnostic Python Detection
**Finding Mac Binaries on Linux**

Normally, `runner.go` should verify the OS (`runtime.GOOS`) before deciding where to look for the Python interpreter. However, to support analyzing macOS games on Linux (detecting version, etc.), we removed the OS guard.

**Location:** `runner/runner.go` -> `findPython`
**The Hack:**
```go
// Check for macOS App Bundle structure (Check unconditionally to support cross-OS inspection)
macExePath := filepath.Join(filepath.Dir(filepath.Dir(libDir)), "MacOS", "python")
if _, err := os.Stat(macExePath); err == nil {
    return macExePath, libDir, true, nil
}
```
We check for the Mac binary *unconditionally*. While we can't *execute* this binary on Linux/Windows, detecting it allows the `GameInfo` struct to be populated correctly, letting the tool report the Ren'Py version and file counts even if it can't run the decompression.

## 3. Pipeline State Mutation
**The `-e -d` Pipeline Fix**

When running `unren-go -e -d`, the tool extracts RPAs and then immediately tries to decompile RPYCs. The problem is that the `game` struct is populated *before* extraction, so the decompiler doesn't know about the files that were just created.

**Location:** `main.go` -> `handleExtractRPA`
**The Hack:**
```go
// Refresh RPYC file list in case new files were extracted
// This ensures that if decompilation runs after this, it sees the new files
if found, err := utils.FindFilesWithExtension(game.GameDir, ".rpyc"); err == nil {
    game.RPYCFiles = found
}
```
We actively mutate the `game.RPYCFiles` slice inside the extraction polling function. This couples the detection state with the extraction action, but it's the most efficient way to ensure the pipeline proceeds without a full re-detection loop.

## 4. Vendor Script Patching
**Fixing `rpatool.py` Exit Codes**

The original `rpatool` library returns exit code `0` (Success) even if it encounters errors during batch extraction. This made implementing the `--clean` flag dangerous, as we rely on exit codes to know if it's safe to delete source files.

**Location:** `files/python/rpatool.py` (and `rpatool_py2.py`)
**The Hack:**
We injected error tracking logic into the upstream script:
```python
errors = 0
# ... inside loops ...
except Exception as e:
    errors += 1
# ... at end of file ...
if errors > 0:
    sys.exit(1)
```
Instead of wrapping the python execution in complex stderr parsing (which is brittle), we patched the "vendor" code directly to behave like a standard CLI tool. This ensures `unren-go` receives a clear signal (`exit status 1`) if *anything* goes wrong, protecting user data.
