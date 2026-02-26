package main

import (
	"fmt"
	"log"
	"os"
	"vk-fasting/pkg/cmd"
	"vk-fasting/pkg/db"
	"vk-fasting/pkg/util"
)

func main() {

	if err := util.CreateFilesAndFolders(); err != nil {
		fmt.Println("Error creating files/folders:", err)
		os.Exit(1)
	}

	f := db.Fastings{}

	err := f.ReadFromFile("DATABASES/FASTING/fasting.json")
	if err != nil {
		log.Fatalf("Fatal error: failed to load walkings database: %v", err)
	}
	
	cmd.CommandLine(&f)
}
