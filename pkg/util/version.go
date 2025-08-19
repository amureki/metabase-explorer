package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func getLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/amureki/metabase-explorer/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func compareVersions(current, latest string) bool {
	// Normalize versions by removing 'v' prefix
	currentNorm := strings.TrimPrefix(current, "v")
	latestNorm := strings.TrimPrefix(latest, "v")

	// Handle dev version
	if currentNorm == "dev" {
		return false // Always allow update from dev version
	}

	// Simple string comparison for semantic versions
	// This works for most cases like "1.2.3" vs "1.2.4"
	return currentNorm == latestNorm
}

func HandleUpdateCommand(currentVersion string) {
	fmt.Println("Checking for updates...")

	// Get the latest version from GitHub
	latestVersion, err := getLatestVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check for updates: %v\n", err)
		fmt.Fprintf(os.Stderr, "You can manually update by running:\n")
		fmt.Fprintf(os.Stderr, "curl -sSL https://raw.githubusercontent.com/amureki/metabase-explorer/main/install.sh | bash\n")
		os.Exit(1)
	}

	// Compare with current version
	if compareVersions(currentVersion, latestVersion) {
		fmt.Printf("✓ Already up to date! Current version: %s\n", currentVersion)
		return
	}

	fmt.Printf("Update available: %s → %s\n", currentVersion, latestVersion)
	fmt.Println("Updating mbx to the latest version...")

	// Download and execute the install script
	cmd := exec.Command("bash", "-c", "curl -sSL https://raw.githubusercontent.com/amureki/metabase-explorer/main/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nYou can manually update by running:\n")
		fmt.Fprintf(os.Stderr, "curl -sSL https://raw.githubusercontent.com/amureki/metabase-explorer/main/install.sh | bash\n")
		os.Exit(1)
	}

	fmt.Printf("✓ Update completed successfully! Updated to version %s\n", latestVersion)
}
