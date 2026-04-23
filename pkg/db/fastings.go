package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"vk-fasting/pkg/color"
	"vk-fasting/pkg/config"
	"vk-fasting/pkg/util"
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

func (f *Fastings) ReadFromFile(path string) error {

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", path, err)
	}
	defer file.Close()

	// Read entire file contents
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	// Unmarshal JSON data
	if err := json.Unmarshal(byteValue, f); err != nil {
		return fmt.Errorf("error parsing JSON from file %s: %w", path, err)
	}

	return nil
}

func (f *Fastings) PrintCLI() {
	fmt.Println(color.Cyan + "VK-FASTING 1.0" + color.Reset)
	fmt.Println(color.Cyan + "------------------------" + color.Reset)
	f.PrintAllFasts()
	fmt.Println(color.Cyan + "\n< Add Update Delete Undo Quit >" + color.Reset)
	fmt.Print("=> ")
}

func (f *Fastings) PrintAllFasts() {
	for i, fast := range f.FASTINGS {
		fmt.Printf(
			"%d. ID:%d  Start:%s  End:%s  Duration:%s  Weight:%s kg\n",
			i+1,
			fast.ID,
			fast.START,
			fast.END,
			fast.DURATION,
			fast.WEIGHT,
		)
	}
}

func (f *Fastings) Add() error {

	newFast, err := f.promptFastingInput(Fast{})
	if err != nil {
		return err
	}

	newFast.ID = f.nextID()

	f.FASTINGS = append(f.FASTINGS, newFast)

	return f.saveToFile()
}

func (f *Fastings) Update(id int) error {
    index, err := f.indexOf(id)
    if err != nil {
        return err
    }
    updated, err := f.promptFastingInput(f.FASTINGS[index])
    if err != nil {
        return err
    }
    f.FASTINGS[index] = updated
    return f.saveToFile()
}

func (f *Fastings) Delete(id int) error {
    index, err := f.indexOf(id)
    if err != nil {
        return err
    }
    f.FASTINGS = append(f.FASTINGS[:index], f.FASTINGS[index+1:]...)
    return f.saveToFile()
}

func (f *Fastings) promptFastingInput(oldFast Fast) (Fast, error) {
    type field struct {
        prompt     string
        suggestion string
        dest       *string
    }

    newFast := Fast{ID: oldFast.ID}

    fields := []field{
        {"Start Date", oldFast.START,    &newFast.START},
        {"End Date",   oldFast.END,      &newFast.END},
        {"Duration",   oldFast.DURATION, &newFast.DURATION},
        {"Weight",     oldFast.WEIGHT,   &newFast.WEIGHT},
    }

    for _, f := range fields {
        val, err := util.PromptWithSuggestion(f.prompt, f.suggestion)
        if err != nil {
            return Fast{}, fmt.Errorf("prompting %q: %w", f.prompt, err)
        }
        *f.dest = val
    }

    return newFast, nil
}

func (f *Fastings) nextID() int {

	maxID := 0

	for _, fast := range f.FASTINGS {
		if fast.ID > maxID {
			maxID = fast.ID
		}
	}

	return maxID + 1
}

func (f *Fastings) saveToFile() error {

	// Format JSON
	walks, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}

	// Save
	err = os.WriteFile(config.LocalFile, walks, 0644)
	if err != nil {
		return err
	}

	// Save Backup
	if util.HardDriveMountCheck() {
		err = os.WriteFile(config.BackupFile, walks, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Fastings) Undo() bool {
	if len(f.FASTINGS) == 0 {
		fmt.Println("No fast to undo.")
		return false
	}

	lastFast := f.FASTINGS[len(f.FASTINGS)-1]
	fmt.Println(lastFast)

	answer, err := util.PromptWithSuggestion("Are you sure you want to delete?", "No")
	if err != nil {
		fmt.Print(err)
		return false
	}

	answer = strings.ToLower(answer)

	if answer == "y" || answer == "yes" {
		f.FASTINGS = f.FASTINGS[:len(f.FASTINGS)-1]

		if err := f.saveToFile(); err != nil {
			fmt.Println(color.Red+"Error saving data:"+color.Reset, err)
			return false
		}

		fmt.Println(color.Yellow + "Last fast removed." + color.Reset)
		return true
	}

	fmt.Println("Undo cancelled.")
	return false
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
