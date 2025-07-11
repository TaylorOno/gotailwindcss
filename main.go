package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	installDirectory = ".tailwindcss"
)

var tailwindexec = getExecName()

func main() {
	setErrorLevel()

	// happy case tailwind cli installed and on the path
	if path, err := exec.LookPath(tailwindexec); err == nil {
		run(path)
		return
	}
	slog.Info("tailwindcss not found on path, trying to download")

	// look for an existing installation or download the cli
	path, err := getTailwind()
	if err == nil {
		run(path)
		return
	}

	slog.Error("unable to find or download tailwind", "msg", err)
}

func setErrorLevel() {
	if _, ok := os.LookupEnv("GOTAILWINDCSS_DEBUG"); ok {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.SetLogLoggerLevel(slog.LevelError)
}

func getExecName() string {
	if runtime.GOOS == "windows" {
		return "tailwindcss.exe"
	}

	return "tailwindcss"
}

func getTailwind() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// if tailwind exists us that version
	if tailwindExists(userHome) && tailwindIsCurrent(userHome) {
		return filepath.Join(userHome, installDirectory), nil
	}

	// download the latest tailwind version to $USER_HOME/bin/tailwind
	slog.Info("downloading tailwind cli", "directory", filepath.Join(userHome, installDirectory, tailwindexec))
	return downloadTailwind(userHome)
}

func downloadTailwind(userHome string) (string, error) {
	execDirectory := filepath.Join(userHome, installDirectory)

	// create a bin directory in the user home if it does not already exist
	if !binExists(userHome) {
		if err := os.MkdirAll(execDirectory, 0755); err != nil {
			return "", err
		}
	}

	file, err := os.Create(filepath.Join(execDirectory, tailwindexec))
	if err != nil {
		return "", err
	}
	defer file.Close()

	// download the tailwind standalone cli
	resp, err := http.DefaultClient.Get(downloadURL())
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

	return execDirectory, nil
}

// downloadURL this is either the download url for the specified version or the latest version.
func downloadURL() string {
	operating, arch := getArch()
	if val, ok := os.LookupEnv("TAILWINDCSS_VERSION"); ok {
		// specified version url
		return fmt.Sprintf("https://github.com/tailwindlabs/tailwindcss/releases/download/%s/tailwindcss-%s-%s", val, operating, arch)
	}

	// latest version url
	return fmt.Sprintf("https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-%s-%s", operating, arch)
}

func tailwindExists(home string) bool {
	if _, err := os.Stat(filepath.Join(home, installDirectory, tailwindexec)); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

// tailwindIsCurrent checks the version that is available on the path to determine if an update is necessary.
func tailwindIsCurrent(home string) bool {
	result := &bytes.Buffer{}
	cmd := exec.Command(filepath.Join(home, installDirectory, tailwindexec))
	cmd.Stdout = result
	if err := cmd.Run(); err != nil {
		return false
	}

	line, _, err := bufio.NewReader(result).ReadLine()
	if err != nil {
		return false
	}

	if !strings.Contains(string(line), version()) {
		return false
	}

	return true
}

// version returns the version from the environment variable TAILWINDCSS_VERSION or the latest release.
func version() string {
	if version, ok := os.LookupEnv("TAILWINDCSS_VERSION"); ok {
		return version
	}

	resp, err := http.Get("https://api.github.com/repos/tailwindlabs/tailwindcss/tags")
	if err != nil {
		log.Fatal("unable to get latest version", "error", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal("unable to get latest version", "status", resp.Status)
	}
	defer resp.Body.Close()

	var result TagsResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("unable to get latest version", "error", err)
	}

	if len(result) == 0 {
		log.Fatal("unable to get latest version", "error", "no tags found")
	}

	return result[0].Name
}

func binExists(home string) bool {
	if _, err := os.Stat(filepath.Join(home, installDirectory)); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func run(path string) {
	// add tailwind to the path

	if runtime.GOOS == "windows" {
		os.Setenv("PATH", path+";"+os.Getenv("PATH"))
	} else {
		os.Setenv("PATH", path+":"+os.Getenv("PATH"))
	}

	// shell out and run the tailwind cli command
	command := exec.Command(tailwindexec, os.Args[1:]...)
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

type TagsResponse []TagInfo

type TagInfo struct {
	Name string `json:"name"`
}
