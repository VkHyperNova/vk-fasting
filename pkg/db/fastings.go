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
			"%d. ID:%d  Start:%s  End:%s  Duration:%s  Weight:%s\n",
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

	// Get new walk data
	newFast, err := f.GetUserInput(Fast{})
	if err != nil {
		return err
	}

	// Add unique ID
	newFast.ID = f.NewID()

	// Add
	f.FASTINGS = append(f.FASTINGS, newFast)

	// Save
	err = f.Save()
	if err != nil {
		return err
	}


	return nil
}

func (f *Fastings) Update(id int) error {

	// Invalid IDs Guard
	if id <= 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}

	// Find and Update
	for index, fast := range f.FASTINGS {

		// Find walk
		if fast.ID == id {

			// Get updated fields
			updatedFast, err := f.GetUserInput(fast)
			if err != nil {
				return err
			}

			// Update
			f.FASTINGS[index] = updatedFast

			// Save
			return f.Save()
		}
	}
	return fmt.Errorf("item with ID %d not found", id)
}

func (f *Fastings) Delete(id int) error {

	// Invalid IDs Guard
	if id <= 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}

	// Find and Delete
	for index, fast := range f.FASTINGS {
		if fast.ID == id {

			// Delete
			f.FASTINGS = append((f.FASTINGS)[:index], (f.FASTINGS)[index+1:]...)

			return f.Save()
		}
	}

	return fmt.Errorf("item with ID %d not found", id)
}

func (f *Fastings) GetUserInput(oldFast Fast) (Fast, error) {

	start := util.PromptWithSuggestion("Start Date", oldFast.START)
	end := util.PromptWithSuggestion("End Date", oldFast.END)
	days := util.PromptWithSuggestion("Duration", oldFast.DURATION)
	weight := util.PromptWithSuggestion("Weight", oldFast.WEIGHT)

	return Fast{
		ID:       oldFast.ID,
		START:    start,
		END:      end,
		DURATION: days,
		WEIGHT:   weight,
	}, nil
}

func (f *Fastings) NewID() int {

	maxID := 0

	for _, fast := range f.FASTINGS {
		if fast.ID > maxID {
			maxID = fast.ID
		}
	}

	return maxID + 1
}

func (f *Fastings) Save() error {

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

	lastWalk := f.FASTINGS[len(f.FASTINGS)-1]
	fmt.Println(lastWalk)

	answer := strings.ToLower(util.PromptWithSuggestion("Are you sure you want to delete?", "No"))

	if answer == "y" || answer == "yes" {
		f.FASTINGS = f.FASTINGS[:len(f.FASTINGS)-1]

		if err := f.Save(); err != nil {
			fmt.Println(color.Red+"Error saving data:"+color.Reset, err)
			return false
		}

		fmt.Println(color.Yellow + "Last walk removed." + color.Reset)
		return true
	}

	fmt.Println("Undo cancelled.")
	return false
}
