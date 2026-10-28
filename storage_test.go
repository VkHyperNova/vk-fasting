package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

/* Helpers */

// stubMount replaces mountBackup so tests never touch a real drive.
// It returns a pointer to the number of times it was called.
func stubMount(t *testing.T, err error) *int {
	t.Helper()
	calls := 0
	old := mountBackup
	mountBackup = func() error {
		calls++
		return err
	}
	t.Cleanup(func() { mountBackup = old })
	return &calls
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return data
}

/* writeFileAtomic */

func TestWriteFileAtomic_WritesContentAndPerm(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")

	if err := writeFileAtomic(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := string(readFile(t, path)); got != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("perm = %v, want 0600", info.Mode().Perm())
	}
}

func TestWriteFileAtomic_CreatesMissingDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c", "out.json")

	if err := writeFileAtomic(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := string(readFile(t, path)); got != "x" {
		t.Errorf("content = %q, want %q", got, "x")
	}
}

func TestWriteFileAtomic_OverwritesExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")

	if err := writeFileAtomic(path, []byte("old content that is longer"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := string(readFile(t, path)); got != "new" {
		t.Errorf("content = %q, want %q", got, "new")
	}
}

func TestWriteFileAtomic_LeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := writeFileAtomic(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "out.json" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory should contain only out.json, got %v", names)
	}
}

func TestWriteFileAtomic_FailsWhenParentIsAFile(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := writeFileAtomic(filepath.Join(blocker, "out.json"), []byte("x"), 0o644); err == nil {
		t.Error("expected an error")
	}
}

/* Load */

func TestLoad_MissingFileIsNotAnError(t *testing.T) {
	f := Fastings{}
	if err := f.Load(filepath.Join(t.TempDir(), "nope.json")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.FASTINGS) != 0 {
		t.Errorf("expected empty data, got %+v", f.FASTINGS)
	}
}

func TestLoad_ValidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fasting.json")
	content := `{"fastings":[{"id":1,"start":"S","end":"E","duration":"48h00m","weight":"80"}]}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	f := Fastings{}
	if err := f.Load(path); err != nil {
		t.Fatal(err)
	}
	if len(f.FASTINGS) != 1 {
		t.Fatalf("got %d fasts, want 1", len(f.FASTINGS))
	}
	want := Fast{ID: 1, START: "S", END: "E", DURATION: "48h00m", WEIGHT: "80"}
	if f.FASTINGS[0] != want {
		t.Errorf("got %+v, want %+v", f.FASTINGS[0], want)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fasting.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	f := Fastings{}
	err := f.Load(path)
	if err == nil {
		t.Fatal("expected a parse error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error should mention the file path, got: %v", err)
	}
}

func TestLoad_UnreadablePath(t *testing.T) {
	// Reading a directory fails with something other than "not exist".
	f := Fastings{}
	if err := f.Load(t.TempDir()); err == nil {
		t.Error("expected a read error")
	}
}

/* Save */

func TestSave_WritesLocalFileAndBackup(t *testing.T) {
	useTempStore(t)
	calls := stubMount(t, nil)

	f := sampleFasts()
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}

	local := readFile(t, LocalFile)
	backup := readFile(t, BackupFile)
	if string(local) != string(backup) {
		t.Error("backup content differs from local file")
	}
	if *calls != 1 {
		t.Errorf("mountBackup called %d times, want 1", *calls)
	}

	info, err := os.Stat(LocalFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("local file perm = %v, want 0600", info.Mode().Perm())
	}
}

func TestSave_RoundTrip(t *testing.T) {
	useTempStore(t)
	stubMount(t, nil)

	f := sampleFasts()
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}

	loaded := Fastings{}
	if err := loaded.Load(LocalFile); err != nil {
		t.Fatal(err)
	}
	if len(loaded.FASTINGS) != len(f.FASTINGS) {
		t.Fatalf("got %d fasts, want %d", len(loaded.FASTINGS), len(f.FASTINGS))
	}
	for i := range f.FASTINGS {
		if loaded.FASTINGS[i] != f.FASTINGS[i] {
			t.Errorf("fast %d: got %+v, want %+v", i, loaded.FASTINGS[i], f.FASTINGS[i])
		}
	}
}

func TestSave_JSONShape(t *testing.T) {
	useTempStore(t)
	stubMount(t, nil)

	f := &Fastings{FASTINGS: []Fast{{ID: 1, START: "S", END: "E", DURATION: "D", WEIGHT: "W"}}}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}

	raw := readFile(t, LocalFile)
	if !strings.Contains(string(raw), "\n  ") {
		t.Error("JSON should be indented")
	}

	var decoded map[string][]map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	items := decoded["fastings"]
	if len(items) != 1 {
		t.Fatalf("expected a 'fastings' array with 1 item, got %v", decoded)
	}
	for _, key := range []string{"id", "start", "end", "duration", "weight"} {
		if _, ok := items[0][key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestSave_BackupMountFailureIsOnlyAWarning(t *testing.T) {
	useTempStore(t)
	stubMount(t, errors.New("no drive"))

	f := sampleFasts()
	var err error
	out := captureStdout(t, func() { err = f.Save() })

	if err != nil {
		t.Fatalf("Save should succeed when only the backup fails, got: %v", err)
	}
	if !strings.Contains(out, "Backup failed") || !strings.Contains(out, "no drive") {
		t.Errorf("expected a backup warning, got: %q", out)
	}
	if len(loadSaved(t).FASTINGS) != 2 {
		t.Error("local file should still be written")
	}
	if _, statErr := os.Stat(BackupFile); !os.IsNotExist(statErr) {
		t.Error("backup file should not exist when mounting failed")
	}
}

func TestSave_BackupWriteFailureIsOnlyAWarning(t *testing.T) {
	useTempStore(t)
	stubMount(t, nil)

	// Make the backup's parent directory impossible to create.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	BackupFile = filepath.Join(blocker, "fasting.json")

	f := sampleFasts()
	var err error
	out := captureStdout(t, func() { err = f.Save() })

	if err != nil {
		t.Fatalf("Save should succeed when only the backup fails, got: %v", err)
	}
	if !strings.Contains(out, "Backup failed") {
		t.Errorf("expected a backup warning, got: %q", out)
	}
	if len(loadSaved(t).FASTINGS) != 2 {
		t.Error("local file should still be written")
	}
}

func TestSave_LocalFailureReturnsErrorAndSkipsBackup(t *testing.T) {
	useBrokenStore(t)
	calls := stubMount(t, nil)

	f := sampleFasts()
	err := f.Save()
	if err == nil {
		t.Fatal("expected an error when the local file cannot be written")
	}
	if !strings.Contains(err.Error(), LocalFile) {
		t.Errorf("error should mention the local path, got: %v", err)
	}
	if *calls != 0 {
		t.Errorf("backup should not be attempted after a local failure (mount called %d times)", *calls)
	}
}

/* backup */

func TestBackup_WritesData(t *testing.T) {
	useTempStore(t)
	calls := stubMount(t, nil)

	f := &Fastings{}
	if err := f.backup([]byte("payload")); err != nil {
		t.Fatal(err)
	}
	if got := string(readFile(t, BackupFile)); got != "payload" {
		t.Errorf("backup content = %q, want %q", got, "payload")
	}
	if *calls != 1 {
		t.Errorf("mountBackup called %d times, want 1", *calls)
	}
}

func TestBackup_MountErrorIsWrappedAndSkipsWrite(t *testing.T) {
	useTempStore(t)
	sentinel := errors.New("no drive")
	stubMount(t, sentinel)

	f := &Fastings{}
	err := f.backup([]byte("payload"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the mount error to be wrapped, got: %v", err)
	}
	if !strings.Contains(err.Error(), "mount drive") {
		t.Errorf("error should say 'mount drive', got: %v", err)
	}
	if _, statErr := os.Stat(BackupFile); !os.IsNotExist(statErr) {
		t.Error("nothing should be written when mounting fails")
	}
}
