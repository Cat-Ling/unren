// UnRen-Go: Cross-platform Ren'Py game utility
// Ported from the original UnRen Windows batch script
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unren/unren-go/detector"
	"github.com/unren/unren-go/patcher"
	"github.com/unren/unren-go/runner"
	"github.com/unren/unren-go/utils"
)

const version = "0.0.5"

func main() {
	// Parse command-line flags
	var (
		showVersion bool
		gameDir     string

		// Action flags
		extract   bool
		decompile bool
		console   bool
		quicksave bool
		skip      bool
		rollback  bool
		all       bool
		clean     bool
	)

	// Custom Usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "UnRen-Go v%s\n", version)
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [game_directory]\n\n", filepath.Base(os.Args[0]))

		fmt.Fprintln(os.Stderr, "General Options:")
		fmt.Fprintln(os.Stderr, "  -h, --help       Show this help message")
		fmt.Fprintln(os.Stderr, "  -v, --version    Show version information")

		fmt.Fprintln(os.Stderr, "\nActions:")
		fmt.Fprintln(os.Stderr, "  -e, --extract    Extract RPA packages")
		fmt.Fprintln(os.Stderr, "  -d, --decompile  Decompile RPYC files")
		fmt.Fprintln(os.Stderr, "  -a, --all        Perform ALL actions (extract, decompile, apply all patches)")

		fmt.Fprintln(os.Stderr, "\nPatches:")
		fmt.Fprintln(os.Stderr, "  --console        Enable Developer Console (SHIFT+O) and Menu (SHIFT+D)")
		fmt.Fprintln(os.Stderr, "  --quicksave      Enable Quick Save (F5) and Quick Load (F9)")
		fmt.Fprintln(os.Stderr, "  --skip           Force enable skipping of unseen content")
		fmt.Fprintln(os.Stderr, "  --rollback       Force enable rollback (scroll wheel)")

		fmt.Fprintln(os.Stderr, "\nAdvanced:")
		fmt.Fprintln(os.Stderr, "  --clean          Remove source files (.rpa/.rpyc) after SUCCESSFUL extraction/decompilation")

		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintf(os.Stderr, "  %s -e -d /path/to/game\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s --all .\n", filepath.Base(os.Args[0]))
	}

	// Define flags (support both short and long versions where applicable)
	// We use direct bindings for one version and manual aliasing for the other if needed,
	// or just document them as aliases if using a third-party lib.
	// Since we are using stdlib 'flag', we have to register explicitly.

	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showVersion, "v", false, "Show version (short)")

	flag.BoolVar(&extract, "extract", false, "Extract RPA")
	flag.BoolVar(&extract, "e", false, "Extract RPA (short)")

	flag.BoolVar(&decompile, "decompile", false, "Decompile RPYC")
	flag.BoolVar(&decompile, "d", false, "Decompile RPYC (short)")

	flag.BoolVar(&all, "all", false, "Do everything")
	flag.BoolVar(&all, "a", false, "Do everything (short)")

	flag.BoolVar(&console, "console", false, "Enable console")
	flag.BoolVar(&quicksave, "quicksave", false, "Enable quicksave")
	flag.BoolVar(&skip, "skip", false, "Enable skip")
	flag.BoolVar(&rollback, "rollback", false, "Enable rollback")

	flag.BoolVar(&clean, "clean", false, "Remove source files on success")
	flag.BoolVar(&clean, "c", false, "Remove source files on success (short)")

	flag.Parse()

	if showVersion {
		fmt.Printf("UnRen-Go v%s\n", version)
		return
	}

	// Handle positional argument for game directory
	if flag.NArg() > 0 {
		gameDir = strings.Trim(flag.Arg(0), "\"'")
	} else {
		gameDir = "."
	}

	// Determine if running in automation mode
	automationMode := extract || decompile || console || quicksave || skip || rollback || all

	// Detect game loop
	var game *detector.GameInfo

	for {
		var err error
		game, err = detector.DetectGame(gameDir)
		if err == nil {
			break
		}

		// Validation failed
		if automationMode {
			fmt.Println()
			fmt.Printf("   ! Error: Cannot locate game files, unable to continue.\n")
			fmt.Printf("     Path: %s\n", gameDir)
			fmt.Println()
			return
		}

		utils.ClearTerminal()
		fmt.Println()
		fmt.Printf("   ! Error: Cannot locate game files in: %s\n", gameDir)
		fmt.Println()

		fmt.Println("  Recovery Options:")
		fmt.Println("    1) Browse for game directory")
		fmt.Println("    2) Show Help")
		fmt.Println("    3) Exit")
		fmt.Println()

		option := readInput("Enter number 1-3 (or drag & drop folder): ")
		fmt.Println()

		// Try to treat input as a path first (drag-and-drop support)
		cleanedPath := strings.Trim(option, "\"'")
		if cleanedPath != "" && utils.DirExists(cleanedPath) {
			gameDir = cleanedPath
			continue
		}

		switch option {
		case "1":
			newDir := browseDirectory(gameDir)
			if newDir != "" {
				gameDir = newDir
				continue // Retry detection with new path
			}
			return // User cancelled browsing
		case "2":
			flag.Usage()
			waitForKey()
			return
		case "3":
			return
		default:
			fmt.Printf("   ! Invalid option: %s\n", option)
			fmt.Println("     Please enter 1, 2, 3, or a valid path.")
			continue
		}
	}

	printGameInfo(game)

	if automationMode {
		if all {
			handleAllOptions(game, clean)
			return
		}

		if extract {
			handleExtractRPA(game, clean)
			fmt.Println()
		}
		if decompile {
			handleDecompileRPYC(game, clean)
			fmt.Println()
		}
		if console {
			handleEnableConsole(game)
		}
		if quicksave {
			handleEnableQuickSave(game)
		}
		if skip {
			handleEnableSkip(game)
		}
		if rollback {
			handleEnableRollback(game)
		}
		return
	}

	// Interactive Mode (Legacy)
	// Main menu loop
	for {
		utils.ClearTerminal()
		printBanner()
		printGameInfo(game)
		printMenu()
		option := readInput("Enter number 1-8 (or any other key to Exit): ")
		fmt.Println()
		printSeparator()
		fmt.Println()

		switch option {
		case "1":
			handleExtractRPA(game, false) // Cleaning disabled in interactive mode for safety unless we add an option
		case "2":
			handleDecompileRPYC(game, false)
		case "3":
			handleEnableConsole(game)
		case "4":
			handleEnableQuickSave(game)
		case "5":
			handleEnableSkip(game)
		case "6":
			handleEnableRollback(game)
		case "7":
			handleOptionsGroup1(game)
		case "8":
			handleAllOptions(game, false)
		default:
			return
		}

		fmt.Println()
		printSeparator()
		fmt.Println()
		fmt.Println("  Finished!")
		fmt.Println()

		again := readInput("Enter \"1\" to go back to the menu, or any other key to exit: ")
		if again != "1" {
			return
		}
		fmt.Println()
	}
}

func printBanner() {
	fmt.Println()
	fmt.Println("     __  __      ____               __          __")
	fmt.Println("    / / / /___  / __ \\___  ____    / /_  ____ _/ /_")
	fmt.Println("   / / / / __ \\/ /_/ / _ \\/ __ \\  / __ \\/ __ `/ __/")
	fmt.Println("  / /_/ / / / / _, _/  __/ / / / / /_/ / /_/ / /_")
	fmt.Println("  \\____/_/ /_/_/ |_|\\___/_/ /_(_)_.___/\\__,_/\\__/ v" + version)
	fmt.Println("   Ported from Sam's script @ f95zone.to/members/sam.7899/")
	fmt.Println()
	printSeparator()
	fmt.Println()
}

func printSeparator() {
	fmt.Println("  ----------------------------------------------------")
}

func printGameInfo(game *detector.GameInfo) {
	fmt.Printf("  Game:         %s\n", game.Name)
	if game.RenPyVersion > 0 {
		fmt.Printf("  Ren'Py:       %d.x (", game.RenPyVersion)
		if game.RenPyVersion >= 8 {
			fmt.Print("Python 3")
		} else {
			fmt.Print("Python 2")
		}
		fmt.Println(")")
	}
	fmt.Printf("  RPA Archives: %d found\n", len(game.RPAFiles))
	fmt.Printf("  RPYC Files:   %d found\n", len(game.RPYCFiles))
	fmt.Println()
	printSeparator()
	fmt.Println()
}

func printMenu() {
	fmt.Println("  Available Options:")
	fmt.Println("    1) Extract RPA packages (in game folder)")
	fmt.Println("    2) Decompile rpyc files (in game folder)")
	fmt.Println()
	fmt.Println("    3) Enable Console and Developer Menu")
	fmt.Println("    4) Enable Quick Save and Quick Load")
	fmt.Println("    5) Force enable skipping of unseen content")
	fmt.Println("    6) Force enable rollback (scroll wheel)")
	fmt.Println()
	fmt.Println("    7) Options 3-6")
	fmt.Println("    8) Options 1-6")
	fmt.Println()
}

func handleExtractRPA(game *detector.GameInfo, clean bool) {
	if !game.HasRPAFiles() {
		fmt.Println("  There were no .rpa files to unpack.")
		return
	}

	// Use Python-based extraction via runner (matches original batch script behavior)
	r, err := runner.NewRunner(game)
	if err != nil {
		fmt.Printf("    ! Failed to find Python interpreter: %v\n", err)
		fmt.Println("      RPA extraction requires the game's bundled Python interpreter.")
		return
	}
	defer r.Cleanup()

	fmt.Println("  Extracting RPA archives...")
	fmt.Println()

	totalErrors := 0

	for _, rpaPath := range game.RPAFiles {
		relPath, _ := filepath.Rel(game.GameDir, rpaPath)
		info, _ := os.Stat(rpaPath)
		fmt.Printf("    + Unpacking \"%s\" - %s\n", relPath, utils.FormatBytes(info.Size()))

		if err := r.ExtractRPA(rpaPath); err != nil {
			fmt.Printf("    ! Failed to extract %s: %v\n", relPath, err)
			totalErrors++
		}
	}

	fmt.Println()
	if totalErrors > 0 {
		fmt.Printf("  Extraction complete with %d errors.\n", totalErrors)
	} else {
		fmt.Println("  Extraction complete.")

		if clean {
			fmt.Println()
			fmt.Println("  Cleaning up RPA files...")
			cleaned := 0
			for _, rpaPath := range game.RPAFiles {
				if err := os.Remove(rpaPath); err == nil {
					cleaned++
				} else {
					fmt.Printf("    ! Failed to remove %s: %v\n", filepath.Base(rpaPath), err)
				}
			}
			fmt.Printf("    Removed %d .rpa file(s).\n", cleaned)
		}
	}

	// Refresh RPYC file list in case new files were extracted
	// This ensures that if decompilation runs after this, it sees the new files
	if found, err := utils.FindFilesWithExtension(game.GameDir, ".rpyc"); err == nil {
		game.RPYCFiles = found
	}
}

func handleDecompileRPYC(game *detector.GameInfo, clean bool) {
	if !game.HasRPYCFiles() {
		fmt.Println("  There were no .rpyc files to decompile.")
		return
	}

	fmt.Println("  Setting up decompiler...")

	// Create runner to use embedded Python scripts
	r, err := runner.NewRunner(game)
	if err != nil {
		fmt.Printf("    ! Failed to find Python interpreter: %v\n", err)
		fmt.Println("    ! RPYC decompilation requires the game's bundled Python interpreter.")
		return
	}
	defer r.Cleanup()

	pyVersion := "Python 2"
	if r.IsPython3 {
		pyVersion = "Python 3"
	}
	fmt.Printf("    Using %s\n", pyVersion)
	fmt.Println()

	fmt.Println("  Searching for rpyc files...")
	fmt.Println()

	success, skipped, failed, err := r.DecompileAllRPYC()
	if err != nil {
		fmt.Printf("    ! Decompilation error: %v\n", err)
	}

	fmt.Println()
	// Match batch script summary format
	fmt.Print("  Summary: ")
	parts := []string{}
	if success > 0 {
		parts = append(parts, fmt.Sprintf("(%d) files newly decompiled", success))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("(%d) files failed to decompile", failed))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("(%d) decompiled files already exist", skipped))
	}
	if len(parts) == 0 {
		fmt.Println("No files processed")
	} else {
		for i, p := range parts {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(p)
		}
		fmt.Println()
	}

	if clean && failed == 0 {
		fmt.Println()
		fmt.Println("  Cleaning up RPYC files...")
		cleaned := 0
		for _, rpycPath := range game.RPYCFiles {
			if err := os.Remove(rpycPath); err == nil {
				cleaned++
			} else {
				fmt.Printf("    ! Failed to remove %s: %v\n", filepath.Base(rpycPath), err)
			}
		}
		fmt.Printf("    Removed %d .rpyc file(s).\n", cleaned)
	} else if clean && failed > 0 {
		fmt.Println()
		fmt.Printf("  ! Skipping cleanup because %d files failed to decompile.\n", failed)
	}
}

func handleEnableConsole(game *detector.GameInfo) {
	fmt.Println("  Creating Developer/Console file...")

	p := patcher.New(game.GameDir, nil)
	if err := p.EnableConsole(); err != nil {
		fmt.Printf("    ! Failed: %v\n", err)
		return
	}

	fmt.Println("    + Console: SHIFT+O")
	fmt.Println("    + Dev Menu: SHIFT+D")
}

func handleEnableQuickSave(game *detector.GameInfo) {
	fmt.Println("  Creating Quick Save/Quick Load file...")

	p := patcher.New(game.GameDir, nil)
	if err := p.EnableQuickSave(); err != nil {
		fmt.Printf("    ! Failed: %v\n", err)
		return
	}

	fmt.Println("    Default hotkeys:")
	fmt.Println("    + Quick Save: F5")
	fmt.Println("    + Quick Load: F9")
}

func handleEnableSkip(game *detector.GameInfo) {
	fmt.Println("  Creating skip file...")

	p := patcher.New(game.GameDir, nil)
	if err := p.EnableSkip(); err != nil {
		fmt.Printf("    ! Failed: %v\n", err)
		return
	}

	fmt.Println("    + You can now skip all text using TAB and CTRL keys")
}

func handleEnableRollback(game *detector.GameInfo) {
	fmt.Println("  Creating rollback file...")

	p := patcher.New(game.GameDir, nil)
	if err := p.EnableRollback(); err != nil {
		fmt.Printf("    ! Failed: %v\n", err)
		return
	}

	fmt.Println("    + You can now rollback using the scrollwheel")
}

func handleOptionsGroup1(game *detector.GameInfo) {
	handleEnableConsole(game)
	fmt.Println()
	handleEnableQuickSave(game)
	fmt.Println()
	handleEnableSkip(game)
	fmt.Println()
	handleEnableRollback(game)
}

func handleAllOptions(game *detector.GameInfo, clean bool) {
	handleExtractRPA(game, clean)
	fmt.Println()
	handleDecompileRPYC(game, clean)
	fmt.Println()
	handleOptionsGroup1(game)
}

func readInput(prompt string) string {
	fmt.Print("  " + prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func waitForKey() {
	fmt.Print("            Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

// browseDirectory allows interactive directory selection
func browseDirectory(startPath string) string {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		currentPath, _ = os.Getwd()
	}

	for {
		fmt.Println("  ----------------------------------------------------")
		fmt.Printf("  Browsing: %s\n", currentPath)
		fmt.Println("  ----------------------------------------------------")

		dirs, err := listSubdirectories(currentPath)
		if err != nil {
			fmt.Printf("  Error listing directory: %v\n", err)
			waitForKey()
			return ""
		}

		fmt.Println("  [..]  Go Up")
		fmt.Println("  [.]   SELECT CURRENT DIRECTORY")
		fmt.Println()

		// List directories with numbers
		for i, dir := range dirs {
			fmt.Printf("  [%d]   %s\n", i+1, dir)
		}

		fmt.Println()
		fmt.Println("  Enter number to navigate, 'u' to go Up, 's' to Select, or 'q' to Quit")
		input := readInput("Selection: ")

		switch strings.ToLower(input) {
		case "u", "..":
			currentPath = filepath.Dir(currentPath)
		case "s", ".":
			return currentPath
		case "q":
			return ""
		default:
			// Try to parse number
			var selection int
			_, err := fmt.Sscanf(input, "%d", &selection)
			if err == nil {
				if selection > 0 && selection <= len(dirs) {
					currentPath = filepath.Join(currentPath, dirs[selection-1])
					continue
				}
			}
			fmt.Println("  Invalid selection.")
		}
	}
}

// listSubdirectories returns a list of subdirectory names in the given path
func listSubdirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}
