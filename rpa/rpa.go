// Package rpa implements Ren'Py Archive (RPA) extraction.
// Supports RPA-2.0 and RPA-3.0 archive formats.
package rpa

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// MaxIndexSize limits the index data size to prevent memory exhaustion.
// Most RPA indexes are under 10MB even for large archives.
const MaxIndexSize = 50 * 1024 * 1024 // 50 MB

// BufferSize for streaming file extraction.
const BufferSize = 64 * 1024 // 64 KB

// Archive represents an RPA archive file.
type Archive struct {
	path    string
	version int    // 2 or 3
	offset  int64  // index offset
	key     uint64 // deobfuscation key (RPA-3.0 only)
}

// FileEntry represents a single file entry in the archive.
type FileEntry struct {
	Path   string
	Offset int64
	Length int64
	Prefix []byte // Optional prefix bytes
}

// ExtractResult contains the result of an extraction operation.
type ExtractResult struct {
	Extracted int
	Skipped   int
	Errors    []error
}

// Open opens an RPA archive for reading.
func Open(path string) (*Archive, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	// Read the header line
	reader := bufio.NewReader(f)
	headerLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	archive := &Archive{path: path}

	// Parse header based on version
	headerLine = strings.TrimSpace(headerLine)
	parts := strings.Split(headerLine, " ")

	switch {
	case strings.HasPrefix(headerLine, "RPA-3.0"):
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid RPA-3.0 header: %s", headerLine)
		}
		archive.version = 3

		// Parse offset (hex)
		offset, err := strconv.ParseInt(parts[1], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse offset: %w", err)
		}
		archive.offset = offset

		// Parse key (hex)
		key, err := strconv.ParseUint(parts[2], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}
		archive.key = key

	case strings.HasPrefix(headerLine, "RPA-2.0"):
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid RPA-2.0 header: %s", headerLine)
		}
		archive.version = 2

		// Parse offset (hex)
		offset, err := strconv.ParseInt(parts[1], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse offset: %w", err)
		}
		archive.offset = offset

	default:
		return nil, fmt.Errorf("unsupported archive format: %s", headerLine)
	}

	return archive, nil
}

// Version returns the RPA version (2 or 3).
func (a *Archive) Version() int {
	return a.version
}

// Path returns the archive file path.
func (a *Archive) Path() string {
	return a.path
}

// ReadIndex reads and parses the archive index.
func (a *Archive) ReadIndex() ([]FileEntry, error) {
	f, err := os.Open(a.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	// Get file size to calculate index size
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat archive: %w", err)
	}

	indexSize := stat.Size() - a.offset
	if indexSize <= 0 {
		return nil, fmt.Errorf("invalid index offset")
	}
	if indexSize > MaxIndexSize {
		return nil, fmt.Errorf("index too large (%d bytes > %d max)", indexSize, MaxIndexSize)
	}

	// Seek to index offset
	if _, err := f.Seek(a.offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to index: %w", err)
	}

	// Read compressed index data with size limit
	compressedData := make([]byte, indexSize)
	if _, err := io.ReadFull(f, compressedData); err != nil {
		return nil, fmt.Errorf("failed to read index data: %w", err)
	}

	// Decompress with zlib using LimitedReader
	zlibReader, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zlibReader.Close()

	// Limit decompressed size (indexes typically decompress to ~2-5x size)
	limitReader := io.LimitReader(zlibReader, MaxIndexSize)
	indexData, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress index: %w", err)
	}

	// Free compressed data early
	compressedData = nil

	// Parse the pickle format
	return a.parsePickledIndex(indexData)
}

// parsePickledIndex parses a Python pickle-encoded index.
// This is a simplified parser that handles the common RPA index format.
func (a *Archive) parsePickledIndex(data []byte) ([]FileEntry, error) {
	// Check for pickle protocol marker
	if len(data) == 0 {
		return nil, fmt.Errorf("empty index data")
	}

	offset := 0
	// Protocol 2 starts with 0x80 0x02
	if data[0] == 0x80 && len(data) > 1 {
		offset = 2
	}

	// Parse the pickled data structure
	return a.parsePickleDict(data[offset:])
}

// parsePickleDict parses a pickled dictionary for file entries.
func (a *Archive) parsePickleDict(data []byte) ([]FileEntry, error) {
	entries := make([]FileEntry, 0, 100) // Pre-allocate reasonable capacity

	// Find string markers and extract filenames and tuples
	// This is a heuristic approach for common RPA formats
	i := 0
	dataLen := len(data)
	for i < dataLen {
		// Look for pickle string opcodes
		switch data[i] {
		case 'U': // SHORT_BINSTRING
			if i+1 >= dataLen {
				i++
				continue
			}
			strLen := int(data[i+1])
			if strLen <= 0 || i+2+strLen > dataLen {
				i++
				continue
			}
			filename := string(data[i+2 : i+2+strLen])

			// Check if this looks like a valid filename
			if isValidFilename(filename) {
				// Search for the associated tuple data (offset, length)
				remaining := dataLen - (i + 2 + strLen)
				if remaining > 100 {
					remaining = 100
				}
				entry, found := a.findFileData(data[i+2+strLen:i+2+strLen+remaining], filename)
				if found {
					entries = append(entries, entry)
				}
			}
			i += 2 + strLen

		case 'X': // BINUNICODE
			if i+4 >= dataLen {
				i++
				continue
			}
			strLen := int(binary.LittleEndian.Uint32(data[i+1:]))
			if strLen <= 0 || strLen > 1000 || i+5+strLen > dataLen {
				i++
				continue
			}
			filename := string(data[i+5 : i+5+strLen])

			if isValidFilename(filename) {
				remaining := dataLen - (i + 5 + strLen)
				if remaining > 100 {
					remaining = 100
				}
				entry, found := a.findFileData(data[i+5+strLen:i+5+strLen+remaining], filename)
				if found {
					entries = append(entries, entry)
				}
			}
			i += 5 + strLen

		default:
			i++
		}
	}

	return entries, nil
}

// findFileData attempts to find offset and length data after a filename.
func (a *Archive) findFileData(data []byte, filename string) (FileEntry, bool) {
	entry := FileEntry{Path: filename}
	dataLen := len(data)

	// Look for tuple markers and integer data
	// Common patterns: BININT, BININT1, BININT2, etc.
	for i := 0; i < dataLen && i < 100; i++ {
		switch data[i] {
		case 'J': // BININT (4 bytes signed)
			if i+4 < dataLen {
				val := int64(binary.LittleEndian.Uint32(data[i+1:]))
				if entry.Offset == 0 && val > 0 {
					entry.Offset = a.deobfuscate(val)
				} else if entry.Length == 0 && val > 0 {
					entry.Length = a.deobfuscate(val)
					return entry, true
				}
			}
		case 'K': // BININT1 (1 byte unsigned)
			if i+1 < dataLen {
				val := int64(data[i+1])
				if entry.Offset == 0 && val > 0 {
					entry.Offset = a.deobfuscate(val)
				} else if entry.Length == 0 {
					entry.Length = a.deobfuscate(val)
					if entry.Offset > 0 {
						return entry, true
					}
				}
			}
		case 'M': // BININT2 (2 bytes unsigned)
			if i+2 < dataLen {
				val := int64(binary.LittleEndian.Uint16(data[i+1:]))
				if entry.Offset == 0 && val > 0 {
					entry.Offset = a.deobfuscate(val)
				} else if entry.Length == 0 {
					entry.Length = a.deobfuscate(val)
					if entry.Offset > 0 {
						return entry, true
					}
				}
			}
		}
	}

	return entry, false
}

// deobfuscate applies the RPA-3.0 key to an offset/length value.
func (a *Archive) deobfuscate(val int64) int64 {
	if a.version == 3 && a.key != 0 {
		return val ^ int64(a.key)
	}
	return val
}

// isValidFilename checks if a string looks like a valid archive path.
func isValidFilename(s string) bool {
	if len(s) < 3 || len(s) > 500 {
		return false
	}

	// Must contain a file extension
	if !strings.Contains(s, ".") {
		return false
	}

	// Common archive content extensions
	validExts := []string{
		".png", ".jpg", ".jpeg", ".webp", ".gif",
		".ogg", ".mp3", ".wav", ".opus",
		".rpy", ".rpyc", ".rpym", ".rpymc",
		".ttf", ".otf",
		".txt", ".json", ".yaml", ".yml",
	}

	ext := strings.ToLower(filepath.Ext(s))
	for _, valid := range validExts {
		if ext == valid {
			return true
		}
	}

	return false
}

// ExtractAll extracts all files from the archive to the destination directory.
func (a *Archive) ExtractAll(destDir string) (*ExtractResult, error) {
	entries, err := a.ReadIndex()
	if err != nil {
		return nil, err
	}

	return a.ExtractFiles(entries, destDir)
}

// ExtractFiles extracts specific file entries to the destination directory.
// Uses streaming to avoid loading large files entirely into memory.
func (a *Archive) ExtractFiles(entries []FileEntry, destDir string) (*ExtractResult, error) {
	result := &ExtractResult{}

	f, err := os.Open(a.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	// Get archive file size for validation
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat archive: %w", err)
	}
	archiveSize := stat.Size()

	// Reusable buffer for streaming
	buf := make([]byte, BufferSize)

	for _, entry := range entries {
		// Validate entry
		if entry.Offset < 0 || entry.Length <= 0 {
			result.Errors = append(result.Errors, fmt.Errorf("%s: invalid offset/length", entry.Path))
			continue
		}
		if entry.Offset+entry.Length > archiveSize {
			result.Errors = append(result.Errors, fmt.Errorf("%s: entry extends beyond archive", entry.Path))
			continue
		}

		outPath := filepath.Join(destDir, entry.Path)

		// Security: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(outPath), filepath.Clean(destDir)) {
			result.Errors = append(result.Errors, fmt.Errorf("%s: path traversal detected", entry.Path))
			continue
		}

		// Check if file already exists
		if _, err := os.Stat(outPath); err == nil {
			result.Skipped++
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", entry.Path, err))
			continue
		}

		// Extract using streaming
		if err := a.extractFileStreaming(f, entry, outPath, buf); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", entry.Path, err))
			continue
		}

		result.Extracted++
	}

	return result, nil
}

// extractFileStreaming extracts a single file using streaming to minimize memory usage.
func (a *Archive) extractFileStreaming(src *os.File, entry FileEntry, destPath string, buf []byte) error {
	// Seek to file position
	if _, err := src.Seek(entry.Offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek error: %w", err)
	}

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create error: %w", err)
	}
	defer dst.Close()

	// Use buffered writer for better performance
	writer := bufio.NewWriter(dst)
	defer writer.Flush()

	// Write prefix if present
	if len(entry.Prefix) > 0 {
		if _, err := writer.Write(entry.Prefix); err != nil {
			return fmt.Errorf("write prefix error: %w", err)
		}
	}

	// Stream copy the file data
	remaining := entry.Length
	if len(entry.Prefix) > 0 {
		remaining -= int64(len(entry.Prefix))
	}

	for remaining > 0 {
		toRead := int64(len(buf))
		if toRead > remaining {
			toRead = remaining
		}

		n, err := src.Read(buf[:toRead])
		if err != nil && err != io.EOF {
			return fmt.Errorf("read error: %w", err)
		}
		if n == 0 {
			break
		}

		if _, err := writer.Write(buf[:n]); err != nil {
			return fmt.Errorf("write error: %w", err)
		}

		remaining -= int64(n)
	}

	return nil
}
