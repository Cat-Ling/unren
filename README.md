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
Download from the releases page for your platform:
- `unren-go` (Linux)
- `unren-go.exe` (Windows)
- `unren-go-darwin` (macOS)

## Usage

### Interactive Mode
```bash
# From game's root directory
./unren-go

# Or specify directory
./unren-go --dir /path/to/game
```

### Command Line
```bash
./unren-go --version    # Show version
./unren-go --no-menu    # Exit after single operation
```

## Menu Options

1. Extract RPA packages
2. Decompile rpyc files
3. Enable Console and Developer Menu
4. Enable Quick Save and Quick Load
5. Force enable skipping of unseen content
6. Force enable rollback (scroll wheel)
7. Options 3-6
8. Options 1, 3-6

## Project Structure

```
unren-go/
├── main.go           # CLI entry point
├── files/            # Embedded RPY templates
├── rpa/              # RPA archive extraction
├── patcher/          # RPY file generation
├── detector/         # Game detection
└── utils/            # Cross-platform utilities
```

## License

MIT License

## Credits

- Original UnRen by [F95Zone community](https://f95zone.to/)
- Go port by the UnRen-Go contributors
