package util

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"vk-fasting/pkg/color"
	"vk-fasting/pkg/config"

	"github.com/peterh/liner"
)

func HardDriveMountCheck() bool {
	if runtime.GOOS != "linux" {
		fmt.Println("This program only works on Linux.")
		return false
	}

	mountPoint := "/media/veikko/VK\\040DATA" // match /proc/mounts format

	file, err := os.Open("/proc/mounts")
	if err != nil {
		fmt.Println("Cannot open /proc/mounts:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == mountPoint {
			fmt.Println("VK-DATA is Mounted!")
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning /proc/mounts:", err)
		return false
	}

	fmt.Println(color.Red + "\nVK DATA is NOT mounted" + color.Reset)
	return false
}

func Input(prompt string) string {

	line := liner.NewLiner()
	defer line.Close()

	userInput, err := line.Prompt(prompt)
	if err != nil {
		panic(err)
	}
	return userInput
}

func ensureFile(path string, content string) error {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("error creating directory for %s: %w", path, err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("error creating file %s: %w", path, err)
		}
	}

	return nil
}

func CreateFilesAndFolders() error {
	

	if err := ensureFile(config.LocalFile, config.DefaultContent); err != nil {
		return err
	}

	if !HardDriveMountCheck() {
		input := Input("Do you want to continue? (y/n) ")
		if strings.ToLower(strings.TrimSpace(input)) != "y" {
			fmt.Println("Exiting program.")
			os.Exit(0)
		}
	} else {
		if err := ensureFile(config.BackupFile, config.DefaultContent); err != nil {
			return err
		}
	}

	return nil
}

func ReadCommand() (string, int, bool) {
	reader := bufio.NewReader(os.Stdin)

	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error:", err)
		return "", 0, false
	}

	line = strings.TrimSpace(line)
	parts := strings.Fields(line)

	if len(parts) == 0 {
		return "", 0, false
	}

	command := strings.ToLower(parts[0])
	id := 0

	if len(parts) > 1 {
		if _, err := fmt.Sscan(parts[1], &id); err != nil {
			fmt.Println("Invalid ID")
			return "", 0, false
		}
	}

	return command, id, true
}

func PressAnyKey() {
	fmt.Print()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}

func ClearScreen() {

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error clearing screen:", err)
	}
}

func PromptWithSuggestion(name string, suggestion string) string {

	line := liner.NewLiner()
	defer line.Close()

	input, err := line.PromptWithSuggestion("   "+name+": ", suggestion, -1)
	if err != nil {
		panic(err)
	}

	return input
}