package main

import (
	"vk-fasting/pkg/cmd"
	"vk-fasting/pkg/db"
	"vk-fasting/pkg/util"
)

func main() {
	// It initializes a Quotes database
	fasting := db.Fasting{}

	// Ensures the necessary directory structure exists
	util.CreateDatabase("FASTING", "fasting.json")

	// Loads quotes from a file
	fasting.ReadFromFile("./FASTING/fasting.json")

	// and starts the command-line interface for user interaction.
	cmd.CommandLine(&fasting)
}
