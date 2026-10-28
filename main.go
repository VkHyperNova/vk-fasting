package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"

	timeLayout      = "02.01.06"
	driveLabel      = "VK DATA"
	driveMountPoint = "/media/veikko/VK DATA"
	file            = "fasting.json"
	baseDB          = "FASTING"
	baseLocal       = "DEVELOPMENT"
	// baseLocal       = "DATABASES"

	usage = `Usage: fasting [command] [id]

Commands:
  a, add          Add a new fast
  u, update <id>  Update the fast with the given ID
  d, delete <id>  Delete the fast with the given ID
  undo            Delete the most recent fast
  help, -h        Show this help

With no command, the last fasts are printed.
`
)

var (
	LocalFile  = filepath.Join(baseLocal, baseDB, file)
	BackupFile = filepath.Join(driveMountPoint, baseLocal, baseDB, file)
	stdin      = bufio.NewReader(os.Stdin)
)

func main() {

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			fmt.Print(usage)
			return
		}
	}

	f := Fastings{}
	if err := f.Load(LocalFile); err != nil {
		log.Fatalf("load from file: %v", err)
	}

	var err error
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "a", "add":
			err = f.Add()
		case "u", "update":
			err = f.Update(os.Args[2])
		case "d", "delete":
			err = f.Delete(os.Args[2])
		case "undo":
			err = f.Undo()
		default:
			fmt.Printf("%serror: unknown command: %s%s\n\n", Red, os.Args[1], Reset)
			fmt.Print(usage)
			os.Exit(1)
		}
	}

	f.Report()

	if err != nil {
		fmt.Printf("%serror: %v%s\n", Red, err, Reset)
		os.Exit(1)
	}

}
