#!/bin/bash
# Cross-platform build script for UnRen-Go
# Builds for Windows, Linux, macOS on x64/x86/ARM architectures

VERSION="0.0.1"
OUTPUT_DIR="dist"
BINARY_NAME="unren-go"

mkdir -p "$OUTPUT_DIR"

echo "Building UnRen-Go v$VERSION for all platforms..."
echo ""

# Build targets: GOOS/GOARCH combinations
# Format: "os/arch[:output_suffix]"
TARGETS=(
    # Windows
    "windows/amd64:windows-x64.exe"
    "windows/386:windows-x86.exe"
    "windows/arm64:windows-arm64.exe"
    
    # Linux
    "linux/amd64:linux-x64"
    "linux/386:linux-x86"
    "linux/arm64:linux-arm64"
    "linux/arm:linux-arm"  # For Raspberry Pi etc.
    
    # macOS
    "darwin/amd64:macos-x64"
    "darwin/arm64:macos-arm64"  # Apple Silicon
)

for target in "${TARGETS[@]}"; do
    IFS=':' read -r platform suffix <<< "$target"
    IFS='/' read -r goos goarch <<< "$platform"
    
    output="$OUTPUT_DIR/${BINARY_NAME}-${suffix}"
    
    echo "  Building $output..."
    
    CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -ldflags="-s -w" -o "$output" . 2>&1
    
    if [ $? -eq 0 ]; then
        size=$(ls -lh "$output" | awk '{print $5}')
        echo "    ✓ $output ($size)"
    else
        echo "    ✗ Failed to build $output"
    fi
done

echo ""
echo "Build complete! Binaries in $OUTPUT_DIR/"
ls -la "$OUTPUT_DIR/"
