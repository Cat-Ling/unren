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
```bash
# From game's root directory
./unren-go

# Or specify directory
./unren-go --dir /path/to/game
```

### Command Line
```bash
./unren-go --no-menu    # Exit after single operation
```

## License

MIT License

## Credits

- [Sam](https://f95zone.to/members/sam.7899/) from F95Zone Community.
