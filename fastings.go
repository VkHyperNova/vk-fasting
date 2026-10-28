package main

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Fast struct {
	ID       int    `json:"id"`
	START    string `json:"start"`
	END      string `json:"end"`
	DURATION string `json:"duration"`
	WEIGHT   string `json:"weight"`
}

type Fastings struct {
	FASTINGS []Fast `json:"fastings"`
}

/* Main Commands */

func (f *Fastings) Add() error {
	newFast, err := f.promptFastingInput(Fast{})
	if err != nil {
		return err
	}
	newFast.ID = f.nextID()

	prev := slices.Clone(f.FASTINGS)
	f.FASTINGS = append(f.FASTINGS, newFast)
	return f.commit(prev)
}

func (f *Fastings) Update(idStr string) error {
	id, err := parseID(idStr)
	if err != nil {
		return err
	}

	index, err := f.indexOf(id)
	if err != nil {
		return err
	}
	updated, err := f.promptFastingInput(f.FASTINGS[index])
	if err != nil {
		return err
	}

	prev := slices.Clone(f.FASTINGS)
	f.FASTINGS[index] = updated
	return f.commit(prev)
}

func (f *Fastings) Delete(idStr string) error {
	id, err := parseID(idStr)
	if err != nil {
		return err
	}
	return f.deleteByID(id)
}

func (f *Fastings) Undo() error {
	if len(f.FASTINGS) == 0 {
		return errors.New("no fast to undo")
	}
	return f.deleteByID(f.FASTINGS[len(f.FASTINGS)-1].ID)
}

/* Helper Functions */

func (f *Fastings) nextID() int {

	maxID := 0

	for _, fast := range f.FASTINGS {
		if fast.ID > maxID {
			maxID = fast.ID
		}
	}

	return maxID + 1
}

func (f *Fastings) indexOf(id int) (int, error) {
	if id <= 0 {
		return -1, fmt.Errorf("invalid ID: %d", id)
	}
	for i, fast := range f.FASTINGS {
		if fast.ID == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("item with ID %d not found", id)
}

func confirm(question string) (bool, error) {
	fmt.Printf("%s (y/N): ", question)
	answer, err := stdin.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("reading input: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true, nil
	}
	return false, nil
}

func printFast(x Fast) {
	fmt.Printf("  ID: %d\n  Start: %s\n  End: %s\n  Duration: %s\n  Weight: %s\n\n",
		x.ID, x.START, x.END, x.DURATION, x.WEIGHT)
}

func (f *Fastings) commit(prev []Fast) error {
	if err := f.Save(); err != nil {
		f.FASTINGS = prev
		return fmt.Errorf("saving data: %w", err)
	}
	return nil
}

func promptField(label, def string) (string, error) {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, err := stdin.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

func promptTime(label, def string, required bool) (string, error) {
	for {
		val, err := promptField(label+" (YYYY-MM-DD HH:MM)", def)
		if err != nil {
			return "", err
		}
		if val == "" {
			if required {
				fmt.Println("This field is required.")
				continue
			}
			return "", nil
		}
		if _, err := time.ParseInLocation(timeLayout, val, time.Local); err != nil {
			fmt.Println("Invalid format, use YYYY-MM-DD HH:MM.")
			continue
		}
		return val, nil
	}
}

func (f *Fastings) promptFastingInput(cur Fast) (Fast, error) {
	if cur.START == "" {
		cur.START = time.Now().Format(timeLayout)
	}

	for {
		start, err := promptTime("Start", cur.START, true)
		if err != nil {
			return Fast{}, err
		}
		end, err := promptTime("End", cur.END, false)
		if err != nil {
			return Fast{}, err
		}

		cur.START, cur.END, cur.DURATION = start, end, ""
		if end != "" {
			s, _ := time.ParseInLocation(timeLayout, start, time.Local)
			e, _ := time.ParseInLocation(timeLayout, end, time.Local)
			if !e.After(s) {
				fmt.Println("End must be after start.")
				continue
			}
			d := e.Sub(s)
			cur.DURATION = fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
		}
		break
	}

	weight, err := promptField("Weight", cur.WEIGHT)
	if err != nil {
		return Fast{}, err
	}
	cur.WEIGHT = weight

	return cur, nil
}

func parseID(s string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("invalid id %q: must be a number", s)
	}
	return id, nil
}

func (f *Fastings) deleteByID(id int) error {
	index, err := f.indexOf(id)
	if err != nil {
		return err
	}

	fmt.Println("Fast to delete:")
	printFast(f.FASTINGS[index])

	ok, err := confirm("Delete this fast?")
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("delete canceled")
	}

	prev := slices.Clone(f.FASTINGS)
	f.FASTINGS = slices.Delete(f.FASTINGS, index, index+1)
	return f.commit(prev)
}
