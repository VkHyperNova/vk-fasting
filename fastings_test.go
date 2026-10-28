package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

/* Helpers */

// Two dates exactly 48h apart, at midnight, mid-year (no DST edge cases).
// They work with both the date-only and the date+time timeLayout.
var (
	day1 = time.Date(2026, 6, 10, 0, 0, 0, 0, time.Local)
	day3 = day1.AddDate(0, 0, 2)
	day5 = day1.AddDate(0, 0, 4)
)

func ts(t time.Time) string { return t.Format(timeLayout) }

// withInput replaces the shared stdin reader for the duration of the test.
func withInput(t *testing.T, input string) {
	t.Helper()
	old := stdin
	stdin = bufio.NewReader(strings.NewReader(input))
	t.Cleanup(func() { stdin = old })
}

// useTempStore points LocalFile/BackupFile at a temp dir so tests never
// touch your real data.
func useTempStore(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	oldLocal, oldBackup := LocalFile, BackupFile
	LocalFile = filepath.Join(dir, "fasting.json")
	BackupFile = filepath.Join(dir, "backup", "fasting.json")
	t.Cleanup(func() { LocalFile, BackupFile = oldLocal, oldBackup })
}

// useBrokenStore makes every save fail: the parent of LocalFile is a regular file.
func useBrokenStore(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldLocal, oldBackup := LocalFile, BackupFile
	LocalFile = filepath.Join(blocker, "fasting.json")
	BackupFile = filepath.Join(dir, "backup", "fasting.json")
	t.Cleanup(func() { LocalFile, BackupFile = oldLocal, oldBackup })
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

// sampleFasts has a gap in the IDs (1 and 3) on purpose.
func sampleFasts() *Fastings {
	return &Fastings{FASTINGS: []Fast{
		{ID: 1, START: ts(day1), END: ts(day3), DURATION: "48h00m", WEIGHT: "80"},
		{ID: 3, START: ts(day3), END: ts(day5), DURATION: "48h00m", WEIGHT: "79"},
	}}
}

func loadSaved(t *testing.T) Fastings {
	t.Helper()
	var g Fastings
	if err := g.Load(LocalFile); err != nil {
		t.Fatalf("loading saved file: %v", err)
	}
	return g
}

/* nextID / indexOf / parseID */

func TestNextID(t *testing.T) {
	if got := (&Fastings{}).nextID(); got != 1 {
		t.Errorf("empty list: got %d, want 1", got)
	}
	if got := sampleFasts().nextID(); got != 4 {
		t.Errorf("with IDs 1 and 3: got %d, want 4", got)
	}
}

func TestIndexOf(t *testing.T) {
	f := sampleFasts()

	if i, err := f.indexOf(3); err != nil || i != 1 {
		t.Errorf("indexOf(3) = %d, %v; want 1, nil", i, err)
	}
	for _, id := range []int{2, 99, 0, -1} {
		if _, err := f.indexOf(id); err == nil {
			t.Errorf("indexOf(%d): expected an error", id)
		}
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"5", 5, false},
		{" 7 \n", 7, false},
		{"-3", -3, false}, // parses fine; indexOf rejects it later
		{"abc", 0, true},
		{"", 0, true},
		{"1.5", 0, true},
	}
	for _, tt := range tests {
		got, err := parseID(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseID(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseID(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

/* Prompts */

func TestConfirm(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"y\n", true},
		{"yes\n", true},
		{"Y\n", true},
		{"  YES  \n", true},
		{"n\n", false},
		{"\n", false},
		{"maybe\n", false},
	}
	for _, tt := range tests {
		withInput(t, tt.in)
		got, err := confirm("Sure?")
		if err != nil {
			t.Errorf("confirm(%q): unexpected error %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("confirm(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestConfirm_ReadError(t *testing.T) {
	withInput(t, "") // immediate EOF
	if _, err := confirm("Sure?"); err == nil {
		t.Error("expected an error on EOF")
	}
}

func TestPromptField(t *testing.T) {
	withInput(t, "hello\n")
	if got, err := promptField("Label", "default"); err != nil || got != "hello" {
		t.Errorf("got %q, %v; want %q", got, err, "hello")
	}

	withInput(t, "\n")
	if got, err := promptField("Label", "default"); err != nil || got != "default" {
		t.Errorf("empty input should keep default; got %q, %v", got, err)
	}

	withInput(t, "")
	if _, err := promptField("Label", ""); err == nil {
		t.Error("expected an error on EOF")
	}
}

func TestPromptTime(t *testing.T) {
	valid := ts(day1)

	t.Run("valid input", func(t *testing.T) {
		withInput(t, valid+"\n")
		got, err := promptTime("Start", "", true)
		if err != nil || got != valid {
			t.Errorf("got %q, %v; want %q", got, err, valid)
		}
	})

	t.Run("invalid then valid", func(t *testing.T) {
		withInput(t, "garbage\n"+valid+"\n")
		got, err := promptTime("Start", "", true)
		if err != nil || got != valid {
			t.Errorf("got %q, %v; want %q", got, err, valid)
		}
	})

	t.Run("empty required then valid", func(t *testing.T) {
		withInput(t, "\n"+valid+"\n")
		got, err := promptTime("Start", "", true)
		if err != nil || got != valid {
			t.Errorf("got %q, %v; want %q", got, err, valid)
		}
	})

	t.Run("empty optional", func(t *testing.T) {
		withInput(t, "\n")
		got, err := promptTime("End", "", false)
		if err != nil || got != "" {
			t.Errorf("got %q, %v; want empty", got, err)
		}
	})

	t.Run("enter keeps default", func(t *testing.T) {
		withInput(t, "\n")
		got, err := promptTime("Start", valid, true)
		if err != nil || got != valid {
			t.Errorf("got %q, %v; want %q", got, err, valid)
		}
	})

	t.Run("EOF", func(t *testing.T) {
		withInput(t, "")
		if _, err := promptTime("Start", "", true); err == nil {
			t.Error("expected an error on EOF")
		}
	})
}

func TestPromptFastingInput_New(t *testing.T) {
	withInput(t, ts(day1)+"\n"+ts(day3)+"\n80.5\n")

	got, err := (&Fastings{}).promptFastingInput(Fast{})
	if err != nil {
		t.Fatal(err)
	}
	if got.START != ts(day1) || got.END != ts(day3) {
		t.Errorf("start/end = %q/%q", got.START, got.END)
	}
	if got.DURATION != "48h00m" {
		t.Errorf("duration = %q, want 48h00m", got.DURATION)
	}
	if got.WEIGHT != "80.5" {
		t.Errorf("weight = %q, want 80.5", got.WEIGHT)
	}
}

func TestPromptFastingInput_Ongoing(t *testing.T) {
	withInput(t, ts(day1)+"\n\n\n") // start, no end, no weight

	got, err := (&Fastings{}).promptFastingInput(Fast{})
	if err != nil {
		t.Fatal(err)
	}
	if got.END != "" || got.DURATION != "" {
		t.Errorf("ongoing fast should have no end/duration, got %q/%q", got.END, got.DURATION)
	}
}

func TestPromptFastingInput_DefaultStartIsNow(t *testing.T) {
	withInput(t, "\n\n\n") // accept everything

	got, err := (&Fastings{}).promptFastingInput(Fast{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := time.ParseInLocation(timeLayout, got.START, time.Local); err != nil {
		t.Errorf("default start %q does not parse: %v", got.START, err)
	}
}

func TestPromptFastingInput_KeepsExistingValues(t *testing.T) {
	withInput(t, "\n\n\n") // Enter on start, end, weight

	cur := Fast{ID: 7, START: ts(day1), END: ts(day3), DURATION: "old", WEIGHT: "81"}
	got, err := (&Fastings{}).promptFastingInput(cur)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 7 || got.START != cur.START || got.END != cur.END || got.WEIGHT != "81" {
		t.Errorf("existing values not kept: %+v", got)
	}
	if got.DURATION != "48h00m" {
		t.Errorf("duration should be recomputed, got %q", got.DURATION)
	}
}

func TestPromptFastingInput_EndBeforeStartAsksAgain(t *testing.T) {
	input := ts(day3) + "\n" + // start
		ts(day1) + "\n" + // end before start -> rejected
		"\n" + // start again: keep
		ts(day5) + "\n" + // valid end
		"75\n" // weight
	withInput(t, input)

	got, err := (&Fastings{}).promptFastingInput(Fast{})
	if err != nil {
		t.Fatal(err)
	}
	if got.START != ts(day3) || got.END != ts(day5) || got.DURATION != "48h00m" || got.WEIGHT != "75" {
		t.Errorf("unexpected result: %+v", got)
	}
}

/* printFast */

func TestPrintFast(t *testing.T) {
	out := captureStdout(t, func() {
		printFast(Fast{ID: 9, START: "S", END: "E", DURATION: "D", WEIGHT: "W"})
	})
	for _, want := range []string{"ID: 9", "Start: S", "End: E", "Duration: D", "Weight: W"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

/* Commands */

func TestAdd(t *testing.T) {
	useTempStore(t)
	f := &Fastings{}

	withInput(t, ts(day1)+"\n"+ts(day3)+"\n80\n")
	if err := f.Add(); err != nil {
		t.Fatal(err)
	}
	if len(f.FASTINGS) != 1 || f.FASTINGS[0].ID != 1 {
		t.Fatalf("after first add: %+v", f.FASTINGS)
	}
	if f.FASTINGS[0].DURATION != "48h00m" || f.FASTINGS[0].WEIGHT != "80" {
		t.Errorf("unexpected fast: %+v", f.FASTINGS[0])
	}

	withInput(t, ts(day3)+"\n\n\n") // ongoing fast
	if err := f.Add(); err != nil {
		t.Fatal(err)
	}
	if len(f.FASTINGS) != 2 || f.FASTINGS[1].ID != 2 {
		t.Fatalf("after second add: %+v", f.FASTINGS)
	}

	saved := loadSaved(t)
	if len(saved.FASTINGS) != 2 {
		t.Errorf("saved file has %d fasts, want 2", len(saved.FASTINGS))
	}
}

func TestAdd_RollsBackWhenSaveFails(t *testing.T) {
	useBrokenStore(t)
	f := sampleFasts()

	withInput(t, ts(day1)+"\n"+ts(day3)+"\n80\n")
	if err := f.Add(); err == nil {
		t.Fatal("expected a save error")
	}
	if len(f.FASTINGS) != 2 {
		t.Errorf("memory should be rolled back to 2 fasts, got %d", len(f.FASTINGS))
	}
}

func TestUpdate(t *testing.T) {
	useTempStore(t)
	f := sampleFasts()

	withInput(t, "\n\n72.5\n") // keep start and end, change weight
	if err := f.Update("3"); err != nil {
		t.Fatal(err)
	}
	if f.FASTINGS[1].ID != 3 || f.FASTINGS[1].WEIGHT != "72.5" {
		t.Errorf("unexpected fast: %+v", f.FASTINGS[1])
	}
	if f.FASTINGS[0].WEIGHT != "80" {
		t.Errorf("other fast was modified: %+v", f.FASTINGS[0])
	}

	saved := loadSaved(t)
	if len(saved.FASTINGS) != 2 || saved.FASTINGS[1].WEIGHT != "72.5" {
		t.Errorf("saved data wrong: %+v", saved.FASTINGS)
	}
}

func TestUpdate_Errors(t *testing.T) {
	useTempStore(t)
	f := sampleFasts()

	if err := f.Update("abc"); err == nil {
		t.Error("expected an error for a non-numeric id")
	}
	if err := f.Update("99"); err == nil {
		t.Error("expected an error for an unknown id")
	}
}

func TestDelete(t *testing.T) {
	useTempStore(t)
	f := sampleFasts()

	withInput(t, "y\n")
	if err := f.Delete("1"); err != nil {
		t.Fatal(err)
	}
	if len(f.FASTINGS) != 1 || f.FASTINGS[0].ID != 3 {
		t.Fatalf("unexpected list: %+v", f.FASTINGS)
	}

	saved := loadSaved(t)
	if len(saved.FASTINGS) != 1 || saved.FASTINGS[0].ID != 3 {
		t.Errorf("saved data wrong: %+v", saved.FASTINGS)
	}
}

func TestDelete_Canceled(t *testing.T) {
	useTempStore(t)
	f := sampleFasts()

	withInput(t, "n\n")
	if err := f.Delete("1"); err == nil {
		t.Error("expected a cancel error")
	}
	if len(f.FASTINGS) != 2 {
		t.Errorf("nothing should be deleted, got %d fasts", len(f.FASTINGS))
	}
}

func TestDelete_Errors(t *testing.T) {
	useTempStore(t)
	f := sampleFasts()

	if err := f.Delete("abc"); err == nil {
		t.Error("expected an error for a non-numeric id")
	}
	if err := f.Delete("99"); err == nil {
		t.Error("expected an error for an unknown id")
	}
	if len(f.FASTINGS) != 2 {
		t.Errorf("nothing should be deleted, got %d fasts", len(f.FASTINGS))
	}
}

func TestDelete_RollsBackWhenSaveFails(t *testing.T) {
	useBrokenStore(t)
	f := sampleFasts()

	withInput(t, "y\n")
	if err := f.Delete("1"); err == nil {
		t.Fatal("expected a save error")
	}
	if len(f.FASTINGS) != 2 || f.FASTINGS[0].ID != 1 {
		t.Errorf("memory should be rolled back, got %+v", f.FASTINGS)
	}
}

func TestUndo(t *testing.T) {
	useTempStore(t)
	f := sampleFasts()

	withInput(t, "y\n")
	if err := f.Undo(); err != nil {
		t.Fatal(err)
	}
	if len(f.FASTINGS) != 1 || f.FASTINGS[0].ID != 1 {
		t.Errorf("the last fast (ID 3) should be gone: %+v", f.FASTINGS)
	}
}

func TestUndo_Canceled(t *testing.T) {
	useTempStore(t)
	f := sampleFasts()

	withInput(t, "n\n")
	if err := f.Undo(); err == nil {
		t.Error("expected a cancel error")
	}
	if len(f.FASTINGS) != 2 {
		t.Errorf("nothing should be deleted, got %d fasts", len(f.FASTINGS))
	}
}

func TestUndo_Empty(t *testing.T) {
	f := &Fastings{}
	if err := f.Undo(); err == nil {
		t.Error("expected an error when there is nothing to undo")
	}
}
