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

const version = "0.0.1"

func main() {
	// Parse command-line flags
	var (
		showVersion    = flag.Bool("version", false, "Show version information")
		gameDir        = flag.String("dir", ".", "Game directory path")
		nonInteractive = flag.Bool("no-menu", false, "Exit after single operation")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("UnRen-Go v%s\n", version)
		fmt.Println("Cross-platform Ren'Py game utility")
		return
	}

	// Print banner (matches original batch script)
	printBanner()

	// Detect game
	game, err := detector.DetectGame(*gameDir)
	if err != nil {
		fmt.Println()
		fmt.Printf("   ! Error: Cannot locate game files, unable to continue.\n")
		fmt.Printf("            Are you sure we're in the game's root or game directory?\n")
		fmt.Println()
		waitForKey()
		return
	}

	printGameInfo(game)

	// Main menu loop
	for {
		printMenu()

		option := readInput("Enter number 1-8 (or any other key to Exit): ")
		fmt.Println()
		printSeparator()
		fmt.Println()

		switch option {
		case "1":
			handleExtractRPA(game)
		case "2":
			handleDecompileRPYC(game)
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
			handleAllOptions(game)
		default:
			return
		}

		fmt.Println()
		printSeparator()
		fmt.Println()
		fmt.Println("  Finished!")
		fmt.Println()

		if *nonInteractive {
			return
		}

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
	fmt.Printf("  Game:         %s\n", filepath.Base(game.RootDir))
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

func handleExtractRPA(game *detector.GameInfo) {
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
	}
}

func handleDecompileRPYC(game *detector.GameInfo) {
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

func handleAllOptions(game *detector.GameInfo) {
	handleExtractRPA(game)
	fmt.Println()
	handleDecompileRPYC(game)
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
