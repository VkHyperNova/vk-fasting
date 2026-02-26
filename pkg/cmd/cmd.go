package cmd

import (
	"fmt"
	"vk-fasting/pkg/color"
	"vk-fasting/pkg/db"
	"vk-fasting/pkg/util"
)

func CommandLine(f *db.Fastings) {
	for {
		f.PrintCLI()

		command, id, ok := util.ReadCommand()
		if !ok {
			continue
		}

		switch command {
		case "a", "add":
			if err := f.Add(); err != nil {
				fmt.Println(color.Red+"Error:"+color.Reset, err)
			} else {
				fmt.Println(color.Yellow + "\nItem Added!" + color.Reset)
			}
			util.PressAnyKey()
			util.ClearScreen()
		case "u", "update":
			if err := f.Update(id); err != nil {
				fmt.Println(color.Red+"Error:"+color.Reset, err)
			} else {
				fmt.Println(color.Yellow + "\nItem Updated!" + color.Reset)
			}
			util.PressAnyKey()
			util.ClearScreen()
		case "d", "delete":
			if err := f.Delete(id); err != nil {
				fmt.Println(color.Red+"Error:"+color.Reset, err)
			} else {
				fmt.Printf(color.Yellow + "\n Item Removed!" + color.Reset)
			}
			util.PressAnyKey()
			util.ClearScreen()
		case "undo":
			f.Undo()
			util.ClearScreen()
		case "q", "quit":
			util.ClearScreen()
			return
		default:
			fmt.Println("Unknown command. Try: add, update, delete, quit")
			util.PressAnyKey()
			util.ClearScreen()
		}
	}
}