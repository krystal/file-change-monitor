package monitor

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func (m *Monitor) CheckFile(path string) bool {
	previousHash, existsInCache := m.fileCache[path]

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		// If the file doesn't exist, we'll just track it as missing so if it
		// is created, we will get a restart.
		if existsInCache && previousHash != "missing" {
			m.logger.Printf("%s has been removed but was previously present\n", path)
			return true
		} else {
			m.fileCache[path] = "missing"
			return false
		}
	}

	defer file.Close()

	hasher := sha1.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		// could not calculate the hash of the file, that's quite a problem really
		// we'll assume that it's OK
		m.logger.Printf("cannot determine hash for %s\n", path)
		return false
	}

	// Calculate the current hash of the file on the disk
	currentHash := hex.EncodeToString(hasher.Sum(nil))

	// Check whether for changes
	if existsInCache && currentHash != previousHash {
		// If the file doesn't exist in the cache, we should cache it and
		// move on because this
		m.logger.Printf("detected change to %s\n", path)
		return true
	}

	// Update the cache
	m.fileCache[path] = currentHash
	return false
}

func (m *Monitor) CheckAllFiles() bool {
	for _, path := range m.Options.Paths {
		if m.CheckFile(path) {
			return true
		}
	}

	return false
}
