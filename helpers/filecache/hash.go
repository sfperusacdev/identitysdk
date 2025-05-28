package filecache

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
)

func GetHashedFilePath(filename string) string {
	const levels = 2
	const segmentLength = 2

	hasher := sha1.New()
	hasher.Write([]byte(filename))
	hash := hex.EncodeToString(hasher.Sum(nil))

	parts := make([]string, levels)
	for i := range levels {
		start := i * segmentLength
		end := start + segmentLength
		parts[i] = hash[start:end]
	}

	parts = append(parts, filename)
	return filepath.Join(parts...)
}
