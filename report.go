package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"text/tabwriter"
)

func (f *Fastings) Report() {
	clearScreen()
	fmt.Println(Cyan + "VK-FASTING 1.0" + Reset)
	fmt.Println(Cyan + "------------------------" + Reset)
	f.PrintLastFasts(10)
	fmt.Println(Cyan + "\n< add, update, delete, undo >" + Reset)
}

func (f Fastings) PrintLastFasts(n int) {
	fasts := f.FASTINGS
	if len(fasts) == 0 {
		fmt.Println("No fasts recorded yet.")
		return
	}
	if len(fasts) > n {
		fasts = fasts[len(fasts)-n:]
	}

	fmt.Printf("Last %d fasts\n\n", len(fasts))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tSTART\tEND\tDURATION\tWEIGHT")
	fmt.Fprintln(w, "--\t-----\t---\t--------\t------")
	for _, x := range fasts {
		end := x.END
		if end == "" {
			end = "(ongoing)"
		}
		weight := x.WEIGHT
		if weight == "" {
			weight = "-"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", x.ID, x.START, end, x.DURATION, weight)
	}
	w.Flush()
}

func clearScreen() {

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error clearing screen:", err)
	}
}
