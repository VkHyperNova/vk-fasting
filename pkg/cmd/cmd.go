package cmd

import (
	"fmt"
	"os"
	"vk-fasting/pkg/db"
	"vk-fasting/pkg/util"
)

func CommandLine(fasting *db.Fasting) {
	for {
		fasting.PrintCLI()

		var input string
		fmt.Printf("=> ")
		fmt.Scanln(&input)

		switch input {
		case "a", "add":
			err := fasting.Add()
			if err != nil {
				fmt.Println(err)
			}
			util.ClearScreen()
		case "undo":
			fasting.Undo()
			util.ClearScreen()
		case "q", "quit":
			util.ClearScreen()
			os.Exit(0)
		default:
			util.ClearScreen()
		}
	}
}
