package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	// happy case tailwind cli installed and on the path
	if path, err := exec.LookPath("tailwindcss"); err == nil {
		run(path)
		return
	}
	fmt.Println("tailwind not found on path")

	// look for an existing installation or download the cli
	path, err := getTailwind()
	if err == nil {
		run(path)
		return
	}

	fmt.Printf("unable to find or download tailwind: %s\n", err)
}

func getTailwind() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// if tailwind exists us that version
	if tailwindExists(userHome) {
		return filepath.Join(userHome, "bin", "tailwindcss"), nil
	}

	// download the latest tailwind version to $USER_HOME/bin/tailwind
	fmt.Printf("downloading tailwind cli to %s\n", filepath.Join(userHome, "bin", "tailwindcss"))
	return downloadTailwind(userHome)
}

func downloadTailwind(userHome string) (string, error) {
	// create a bin directory in the user home if it does not already exist
	if !binExists(userHome) {
		if err := os.MkdirAll(filepath.Join(userHome, "bin"), 0755); err != nil {
			return "", err
		}
	}

	file, err := os.Create(filepath.Join(userHome, "bin", "tailwindcss"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	operating, arch := getArch()
	resp, err := http.DefaultClient.Get(fmt.Sprintf("https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-%s-%s", operating, arch))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad response code: %v", resp.Status)
	}

	if _, err = io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	if err = os.Chmod(file.Name(), 0775); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func tailwindExists(home string) bool {
	if _, err := os.Stat(filepath.Join(home, "bin", "tailwindcss")); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func binExists(home string) bool {
	if _, err := os.Stat(filepath.Join(home, "bin")); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func run(path string) {
	// add tailwind to the path
	os.Setenv("PATH", path+":"+os.Getenv("PATH"))

	// shell out and run the tailwind cli command
	command := exec.Command("tailwindcss", os.Args[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		log.Fatal(err)
	}
}

func getArch() (string, string) {
	if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			return "macos", "arm64"
		}
		return "macos", "x64"
	}

	if runtime.GOOS == "linux" {
		if runtime.GOARCH == "arm64" {
			return "linux", "arm64"
		}
		return "linux", "x64"
	}

	return "windows", "x64.exe"
}
