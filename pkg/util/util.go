package util

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"vk-fasting/pkg/config"
)

func CreateDatabase(dirName string, fileName string) {

	// Create local path
	localPath := "./" + dirName + "/" + fileName
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		CreateDirectory(dirName)
		CreateFile(localPath)
		fmt.Println("LOCALPATH CREATED!")
	}

	// Create backup path
	backupPath := "/media/veikko/VK DATA/DATABASES/" + dirName + "/" + fileName
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		CreateDirectory("/media/veikko/VK DATA/DATABASES/" + dirName + "/")
		CreateFile(backupPath)
		fmt.Println("BACKUP PATH CREATED!")
	}
}

func CreateDirectory(dirName string) {
	err := os.Mkdir(dirName, 0700)
	if err != nil {
		fmt.Println(err)
	}
}

func CreateFile(path string) {
	err := os.WriteFile(path, []byte(`{"fasting": []}`), 0644)
	if err != nil {
		panic(err)
	}
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

func isMounted(mountPoint string) (bool, error) {
    file, err := os.Open("/proc/mounts")
    if err != nil {
        return false, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        fields := strings.Fields(line)
        if len(fields) >= 2 && fields[1] == mountPoint {
            return true, nil
        }
    }

    return false, scanner.Err()
}

func IsVKDataMounted() {

	if runtime.GOOS != "linux" {
        fmt.Println("This program only works on Linux.")
        return
    }

	mountPoint := "/media/veikko/VK\\040DATA" // change to your actual mount path

    mounted, err := isMounted(mountPoint)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    if mounted {
        fmt.Println(config.Green + "<< VK DATA is mounted >>" + config.Reset)
    } else {
        fmt.Println(config.Red + "<< VK DATA is NOT mounted >>" + config.Reset)
    }
}

