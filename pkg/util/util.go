package util

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"vk-fasting/pkg/config"

	"github.com/peterh/liner"
)
/* Storage */

const (
	driveLabel      = "VK DATA"
	driveMountPoint = "/media/veikko/VK DATA"
)

func InitLocalStorage() error {
	return ensureFile(config.LocalFile, config.DefaultContent)
}

func InitBackupStorage() error {
	mounted, err := IsDriveMounted()
	if err != nil {
		return fmt.Errorf("mount check failed: %w", err)
	}

	if !mounted {
		input, err := PromptWithSuggestion("Drive not mounted. Try to mount it? (y/n) ", "y")
		if err != nil {
			return err
		}
		if strings.ToLower(strings.TrimSpace(input)) != "y" {
			return nil
		}
		if err := unlockAndMount(); err != nil {
			return fmt.Errorf("failed to mount drive: %w", err)
		}
		if mounted, err = IsDriveMounted(); err != nil || !mounted {
			return fmt.Errorf("drive still not mounted after mount attempt")
		}
		// Program did the mounting
		return ensureFile(config.BackupFile, config.DefaultContent)
	}

	// Drive was already mounted manually
	return ensureFile(config.BackupFile, config.DefaultContent)
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

func IsDriveMounted() (bool, error) {
	if runtime.GOOS != "linux" {
		return false, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	device, err := findDeviceByLabel(driveLabel)
	if err != nil {
		return false, fmt.Errorf("could not resolve device: %w", err)
	}

	file, err := os.Open("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("cannot open /proc/mounts: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 3)
		if len(parts) >= 2 && parts[0] == device {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func unlockAndMount() error {
	device, err := findDeviceByLabel(driveLabel)
	if err != nil {
		return fmt.Errorf("could not find drive: %w", err)
	}
	fmt.Printf("Found drive at %s\n", device)

	if err := mountDevice(device); err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}
	fmt.Printf("Drive mounted at %s\n", driveMountPoint)
	return nil
}

func findDeviceByLabel(label string) (string, error) {
	out, err := exec.Command("lsblk", "-o", "NAME,LABEL", "-r", "-n").Output()
	if err != nil {
		return "", fmt.Errorf("lsblk failed: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && decodeLsblkLabel(fields[1]) == label {
			return "/dev/" + fields[0], nil
		}
	}
	return "", fmt.Errorf("label '%s' not found (is the drive plugged in?)", label)
}

func decodeLsblkLabel(s string) string {
	return strings.NewReplacer(
		`\x20`, " ",
		`\x09`, "\t",
		`\x0a`, "\n",
		`\x5c`, `\`,
	).Replace(s)
}

func mountDevice(device string) error {
	cmd := exec.Command("udisksctl", "mount", "-b", device)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func UnmountDrive() error {
	device, err := findDeviceByLabel(driveLabel)
	if err != nil {
		return fmt.Errorf("could not find drive: %w", err)
	}

	prompt := fmt.Sprintf("Do you want to unmount drive: %s? (y/n) ", driveMountPoint)
	input, err := PromptWithSuggestion(prompt, "n")
	if err != nil {
		return err
	}

	input = strings.ToLower(input)

	if input == "y" || input == "yes" {
		fmt.Println("Unmounting drive...")
		cmd := exec.Command("udisksctl", "unmount", "-b", device)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Warning: failed to unmount drive:", err)
			return err
		}
		fmt.Println("Drive unmounted successfully")
	} else {
		fmt.Println("Unmount canceled!")
	}

	return nil
}

/* Other */ 

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

func PromptWithSuggestion(name string, suggestion string) (string, error) {

	line := liner.NewLiner()
	defer line.Close()

	input, err := line.PromptWithSuggestion("   "+name+": ", suggestion, -1)
	if err != nil {
		return input, err
	}

	return input, nil
}