# UnRen-Go

Cross-platform Ren'Py game utility - Ported from [Unren 1.0.11d](https://f95zone.to/threads/unren-bat-v1-0-11d-rpa-extractor-rpyc-decompiler-console-developer-menu-enabler.3083/) 

## Features

- **RPA Extraction**: Extract files from `.rpa` archives (RPA-2.0 and RPA-3.0)
- **Enable Console**: Developer console (SHIFT+O) and dev menu (SHIFT+D)
- **Quick Save/Load**: F5/F9 hotkeys for quick save and load
- **Skip Unseen**: Force-enable skipping of all text including unseen content
- **Rollback**: Enable infinite rollback with scroll wheel

## Installation

### Build from source
```bash
go build -o unren-go .
```

### Pre-built binaries
Download from the [releases page](https://github.com/Cat-Ling/unren) for your platform:

## Usage

### Interactive Mode
Run without arguments to launch the interactive menu:
```bash
./unren-go
```

If the game files aren't found, you'll see a **Recovery Menu** allowing you to browse directories interactively to locate the game.

### Automation & CLI Mode
Run with flags to perform actions immediately (no menu):
```bash
# Extract RPA and Decompile RPYC
./unren-go -e -d /path/to/game

# Perform ALL actions (Extract + Decompile + Apply all patches)
./unren-go --all /path/to/game

# Enable specific features
./unren-go --console --skip /path/to/game
```

### Options
| Flag | Short | Description |
|------|-------|-------------|
| `--extract` | `-e` | Extract RPA packages |
| `--decompile` | `-d` | Decompile RPYC files |
| `--clean` | `-c` | Cleanup processed source files |
| `--all` | `-a` | Perform all actions |
| `--console` | | Enable Developer Console/Menu |
| `--quicksave` | | Enable Quick Save/Load |
| `--skip` | | Force enable skipping unseen content |
| `--rollback` | | Force enable infinite rollback |
| `--help` | `-h` | Show help and valid usages |
| `--version` | `-v` | Show version info |

## License

MIT License

## Credits

- [Sam](https://f95zone.to/members/sam.7899/) from F95Zone Community.
