package bundles

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
)

// write serialises data to JSON, gzip-compresses it, and writes it
// atomically to destPath (temp file in same directory + rename).
func write(data any, destPath string) error {
	tmp, err := os.CreateTemp(filepath.Dir(destPath), "bundle-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	writeErr := func() error {
		gz := gzip.NewWriter(tmp)
		if err := json.NewEncoder(gz).Encode(data); err != nil {
			return err
		}
		if err := gz.Close(); err != nil {
			return err
		}
		return tmp.Close()
	}()

	if writeErr != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return writeErr
	}

	if err := os.Rename(tmpName, destPath); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	return nil
}
