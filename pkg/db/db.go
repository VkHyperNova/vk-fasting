package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
	"vk-fasting/pkg/config"
	"vk-fasting/pkg/util"
)

type EatingTime struct {
	ID   int       `json:"id"`
	DATE time.Time `json:"date"`
}

type Fasting struct {
	Fasting []EatingTime `json:"fasting"` // Slice containing multiple Quote instances.
}

func (e *Fasting) PrintCLI() {

	fmt.Println(config.Purple + "VK-FASTING" + config.Reset)

	util.IsVKDataMounted()

	e.PrintPreviousTimes()

	e.PrintLastMeal()

}

func (e *Fasting) Undo() error {

	e.Fasting = e.Fasting[:len(e.Fasting)-1]

	err := e.SaveToFile("./FASTING/fasting.json")
	if err != nil {
		return err
	}

	err = e.SaveToFile("/media/veikko/VK DATA/DATABASES/FASTING/fasting.json")
	if err != nil {
		return err
	}

	return nil
}

func (e *Fasting) PrintPreviousTimes() {

	var record time.Duration

	for id := 0; id < len(e.Fasting)-1; id++ {
		firstMeal := e.Fasting[id].DATE.Format("15:04:05")
		secondMeal := e.Fasting[id+1].DATE.Format("15:04:05")
		difference := e.Fasting[id+1].DATE.Sub(e.Fasting[id].DATE)

		if record < difference {
			record = difference
		}
		fmt.Println(firstMeal, " - ", secondMeal, " = ", difference)
	}
	fmt.Println(config.Green,"\nRecord time: ", record, config.Reset)
}

func (e *Fasting) PrintLastMeal() {
	now := time.Now()

	size := len(e.Fasting)

	if size > 0 {
		lastMeal := e.Fasting[size-1]
		fmt.Println("Last Meal: ", lastMeal.DATE.Format("15:04:05"))
		fmt.Printf("Elapsed: %v\n", now.Sub(lastMeal.DATE))
	}
}

func (e *Fasting) GenerateUniqueID() int {

	maxID := 0

	for _, EatingTime := range e.Fasting {
		if EatingTime.ID > maxID {
			maxID = EatingTime.ID
		}
	}

	return maxID + 1
}

func (e *Fasting) ReadFromFile(path string) {

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(byteValue, e)
	if err != nil {
		panic(err)
	}
}

func (e *Fasting) SaveToFile(path string) error {

	byteValue, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, byteValue, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (e *Fasting) Add() error {

	EatingTime := EatingTime{}
	EatingTime.ID = e.GenerateUniqueID()
	EatingTime.DATE = time.Now()

	e.Fasting = append(e.Fasting, EatingTime)

	// windowsPath := "D:\\DATABASES\\FASTING\\fasting.json"

	err := e.SaveToFile("./FASTING/fasting.json")
	if err != nil {
		return err
	}

	err = e.SaveToFile("/media/veikko/VK DATA/DATABASES/FASTING/fasting.json")
	if err != nil {
		return err
	}

	return nil
}
