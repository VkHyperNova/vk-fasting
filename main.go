package main

import (
	"log"
	"vk-fasting/pkg/cmd"
	"vk-fasting/pkg/config"
	"vk-fasting/pkg/db"
	"vk-fasting/pkg/util"
)

func main() {

	if err := util.InitLocalStorage(); err != nil {
		log.Fatalf("Error creating files/folders: %v", err)

	}

	f := db.Fastings{}

	err := f.LoadFromFile(config.LocalFile)
	if err != nil {
		log.Fatalf("Fatal error: failed to load fastings database: %v", err)
	}
	
	cmd.CommandLine(&f)
}
