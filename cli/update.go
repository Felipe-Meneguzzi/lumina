package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

const (
	luminaRepo    = "Felipe-Meneguzzi/lumina"
	ghAPIBase     = "https://api.github.com/repos/" + luminaRepo
	ghReleasesURL = "https://github.com/" + luminaRepo + "/releases/download"
)

type ghRelease struct {
	TagName string `json:"tag_name"`
}

// SelfUpdate checks for a newer release on GitHub and, if found, downloads and
// replaces the running binary. current is the version reported by --version.
func SelfUpdate(current string, out, errOut io.Writer) error {
	fmt.Fprintln(out, "checking for updates...")

	latest, err := fetchLatestTag()
	if err != nil {
		return fmt.Errorf("lumina: could not fetch releases: %w", err)
	}

	if latest == current {
		fmt.Fprintf(out, "lumina %s is already the latest version\n", current)
		return nil
	}

	fmt.Fprintf(out, "new version: %s (installed: %s)\ndownloading...\n", latest, current)

	assetName := fmt.Sprintf("lumina-%s-%s", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("%s/%s/%s", ghReleasesURL, latest, assetName)

	tmp, err := downloadBinary(url)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("lumina: could not determine binary path: %w", err)
	}

	if err := replaceSelf(self, tmp); err != nil {
		return err
	}

	fmt.Fprintf(out, "lumina updated to %s at %s\n", latest, self)
	return nil
}

func fetchLatestTag() (string, error) {
	resp, err := http.Get(ghAPIBase + "/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", fmt.Errorf("unexpected API response: %w", err)
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("no release found in %s", luminaRepo)
	}
	return rel.TagName, nil
}

func downloadBinary(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("lumina: error downloading binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lumina: asset not found (status %d)\n  URL: %s", resp.StatusCode, url)
	}

	f, err := os.CreateTemp("", "lumina-update-*")
	if err != nil {
		return "", fmt.Errorf("lumina: error creating temporary file: %w", err)
	}
	name := f.Name()

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(name)
		return "", fmt.Errorf("lumina: error writing download: %w", err)
	}
	f.Close()

	if err := os.Chmod(name, 0755); err != nil {
		os.Remove(name)
		return "", fmt.Errorf("lumina: error setting permissions: %w", err)
	}
	return name, nil
}

// replaceSelf performs an atomic rename: self → self.old, newBin → self, removes .old.
// On failure of the second rename, rolls back by restoring self.old.
func replaceSelf(self, newBin string) error {
	backup := self + ".old"

	if err := os.Rename(self, backup); err != nil {
		return fmt.Errorf("lumina: could not back up the current binary: %w", err)
	}

	if err := os.Rename(newBin, self); err != nil {
		_ = os.Rename(backup, self) // rollback
		return fmt.Errorf("lumina: could not install new version: %w", err)
	}

	_ = os.Remove(backup)
	return nil
}
