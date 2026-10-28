package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

/* Helpers */

func manyFasts(n int) Fastings {
	f := Fastings{}
	for i := 1; i <= n; i++ {
		f.FASTINGS = append(f.FASTINGS, Fast{
			ID:       i,
			START:    fmt.Sprintf("S%d", i),
			END:      fmt.Sprintf("E%d", i),
			DURATION: "1h00m",
			WEIGHT:   "80",
		})
	}
	return f
}

// rowFor returns the printed table row that starts with the given ID, or "".
func rowFor(out string, id int) string {
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == strconv.Itoa(id) {
			return line
		}
	}
	return ""
}

/* PrintLastFasts */

func TestPrintLastFasts_Empty(t *testing.T) {
	out := captureStdout(t, func() { Fastings{}.PrintLastFasts(10) })

	if !strings.Contains(out, "No fasts recorded yet.") {
		t.Errorf("unexpected output: %q", out)
	}
	if strings.Contains(out, "START") {
		t.Error("table header should not be printed when there are no fasts")
	}
}

func TestPrintLastFasts_FewerThanN(t *testing.T) {
	f := manyFasts(3)
	out := captureStdout(t, func() { f.PrintLastFasts(10) })

	if !strings.Contains(out, "Last 3 fasts") {
		t.Errorf("missing 'Last 3 fasts' title:\n%s", out)
	}
	for _, h := range []string{"ID", "START", "END", "DURATION", "WEIGHT"} {
		if !strings.Contains(out, h) {
			t.Errorf("missing header %q", h)
		}
	}
	for id := 1; id <= 3; id++ {
		if rowFor(out, id) == "" {
			t.Errorf("missing row for ID %d", id)
		}
	}
}

func TestPrintLastFasts_ExactlyN(t *testing.T) {
	f := manyFasts(10)
	out := captureStdout(t, func() { f.PrintLastFasts(10) })

	if !strings.Contains(out, "Last 10 fasts") {
		t.Errorf("missing 'Last 10 fasts' title:\n%s", out)
	}
	for id := 1; id <= 10; id++ {
		if rowFor(out, id) == "" {
			t.Errorf("missing row for ID %d", id)
		}
	}
}

func TestPrintLastFasts_OnlyLastN(t *testing.T) {
	f := manyFasts(12)
	out := captureStdout(t, func() { f.PrintLastFasts(10) })

	if !strings.Contains(out, "Last 10 fasts") {
		t.Errorf("missing 'Last 10 fasts' title:\n%s", out)
	}
	for _, id := range []int{1, 2} {
		if rowFor(out, id) != "" {
			t.Errorf("ID %d is too old and should not be printed", id)
		}
	}
	for id := 3; id <= 12; id++ {
		if rowFor(out, id) == "" {
			t.Errorf("missing row for ID %d", id)
		}
	}
}

func TestPrintLastFasts_KeepsOrder(t *testing.T) {
	f := manyFasts(12)
	out := captureStdout(t, func() { f.PrintLastFasts(10) })

	prev := -1
	for id := 3; id <= 12; id++ {
		pos := strings.Index(out, rowFor(out, id))
		if pos <= prev {
			t.Errorf("ID %d is out of order (oldest of the last N should come first)", id)
		}
		prev = pos
	}
}

func TestPrintLastFasts_OngoingAndEmptyWeight(t *testing.T) {
	f := Fastings{FASTINGS: []Fast{{ID: 1, START: "S1", END: "", DURATION: "", WEIGHT: ""}}}
	out := captureStdout(t, func() { f.PrintLastFasts(10) })

	row := rowFor(out, 1)
	if !strings.Contains(row, "(ongoing)") {
		t.Errorf("empty END should print (ongoing), row: %q", row)
	}
	fields := strings.Fields(row)
	if fields[len(fields)-1] != "-" {
		t.Errorf("empty WEIGHT should print -, row: %q", row)
	}
}

func TestPrintLastFasts_ColumnsAreAligned(t *testing.T) {
	// IDs 3..12 mix one- and two-digit IDs, which is where misalignment shows.
	f := manyFasts(12)
	out := captureStdout(t, func() { f.PrintLastFasts(10) })

	var header string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "ID") {
			header = line
			break
		}
	}
	if header == "" {
		t.Fatalf("header row not found:\n%s", out)
	}
	col := strings.Index(header, "START")

	for _, x := range f.FASTINGS[2:] {
		row := rowFor(out, x.ID)
		if got := strings.Index(row, x.START); got != col {
			t.Errorf("ID %d: START column at %d, header at %d\n%s", x.ID, got, col, out)
		}
	}
}

/* Report */

func TestReport(t *testing.T) {
	t.Setenv("TERM", "xterm") // so `clear` behaves the same everywhere

	f := manyFasts(3)
	out := captureStdout(t, func() { f.Report() })

	for _, want := range []string{
		"VK-FASTING 1.0",
		"------------------------",
		"Last 3 fasts",
		"add, update, delete, undo",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Report output missing %q", want)
		}
	}
}

func TestReport_Empty(t *testing.T) {
	t.Setenv("TERM", "xterm")

	f := Fastings{}
	out := captureStdout(t, func() { f.Report() })

	if !strings.Contains(out, "No fasts recorded yet.") {
		t.Errorf("expected the empty message, got: %q", out)
	}
}

/* clearScreen */

// Smoke test only: the exact output depends on the terminal and on whether a
// `clear` command exists, so this just checks that it doesn't panic.
func TestClearScreen_DoesNotPanic(t *testing.T) {
	t.Setenv("TERM", "xterm")
	captureStdout(t, clearScreen)
}
