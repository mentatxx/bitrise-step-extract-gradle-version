package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

/**
 * Extracts the versionName and versionCode from the given gradle file
 * @param gradleFilePath The path to the gradle file
 * @return The versionName and versionCode
 */

func extractVersionFromGradleFile(gradleFilePath string) (string, string, error) {
	file, err := os.Open(gradleFilePath)
	if err != nil {
		return "", "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error when closing:", err)
		}
	}()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	state := 0
	versionName := ""
	versionCode := ""
	versionNameRegexp := regexp.MustCompile(`versionName\s+'(.*)'`)
	versionCodeRegexp := regexp.MustCompile(`versionCode\s+(\d+)`)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		// Simple state machine. Looking for android -> defaultConfig and then versionName and versionCode
		if state == 0 && strings.Contains(line, "android {") {
			state = 1
		}
		if state == 1 && strings.Contains(line, "defaultConfig {") {
			state = 2
		}
		if state == 2 {
			matches := versionNameRegexp.FindStringSubmatch(line)
			if len(matches) > 1 {
				versionName = matches[1]
			}
			matches = versionCodeRegexp.FindStringSubmatch(line)
			if len(matches) > 1 {
				versionCode = matches[1]
			}
		}
	}
	return versionName, versionCode, nil
}

func main() {
	sourceDir := os.Getenv("SOURCE_DIR")
	buildNumber := os.Getenv("BITRISE_BUILD_NUMBER")
	versionName, versionCode, error := extractVersionFromGradleFile(sourceDir + "/android/app/build.gradle")
	if error != nil {
		fmt.Printf("Failed to extract version from gradle file, error: %#v", error)
		os.Exit(1)
	}

	//
	// --- Step Outputs: Export Environment Variables for other Steps:
	// You can export Environment Variables for other Steps with
	//  envman, which is automatically installed by `bitrise setup`.
	fmt.Printf("REACT_APP_VERSION_NAME = %s\n", versionName)
	fmt.Printf("REACT_APP_VERSION_CODE = %s\n", versionCode)
	fmt.Printf("REACT_APP_BUILD_NUMBER = %s\n", buildNumber)

	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", "REACT_APP_VERSION_NAME", "--value", versionName).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		os.Exit(1)
	}
	cmdLog, err = exec.Command("bitrise", "envman", "add", "--key", "REACT_APP_VERSION_CODE", "--value", versionCode).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		os.Exit(1)
	}
	cmdLog, err = exec.Command("bitrise", "envman", "add", "--key", "REACT_APP_BUILD_NUMBER", "--value", buildNumber).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		os.Exit(1)
	}
	// You can find more usage examples on envman's GitHub page
	//  at: https://github.com/bitrise-io/envman

	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}
