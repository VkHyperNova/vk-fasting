package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/VkHyperNova/vkmount"
)

/* Storage */

var mountBackup = func() error {
	return vkmount.Drive{Label: driveLabel}.Mount()
}

func (f *Fastings) Save() error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding fastings: %w", err)
	}

	if err := writeFileAtomic(LocalFile, data, 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", LocalFile, err)
	}

	if err := f.backup(data); err != nil {
		fmt.Println(Yellow + "Backup failed: " + err.Error() + Reset)
	}

	return nil
}

func (f *Fastings) Load(source string) error {
	byteValue, err := os.ReadFile(source)

	// No file yet: start with empty data
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", source, err)
	}

	if err := json.Unmarshal(byteValue, f); err != nil {
		return fmt.Errorf("error parsing JSON from file %s: %w", source, err)
	}

	return nil
}

/* Helper Functions */

func (f *Fastings) backup(data []byte) error {
	if err := mountBackup(); err != nil {
		return fmt.Errorf("mount drive: %w", err)
	}
	return writeFileAtomic(BackupFile, data, 0o600)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // harmless after a successful rename

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Chmod(tmp.Name(), perm); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replacing %s: %w", path, err)
	}

	return nil
}
