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
	fmt.Fprintln(out, "verificando atualizações...")

	latest, err := fetchLatestTag()
	if err != nil {
		return fmt.Errorf("lumina: não foi possível consultar releases: %w", err)
	}

	if latest == current {
		fmt.Fprintf(out, "lumina %s já é a versão mais recente\n", current)
		return nil
	}

	fmt.Fprintf(out, "nova versão: %s (instalada: %s)\nbaixando...\n", latest, current)

	assetName := fmt.Sprintf("lumina-%s-%s", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("%s/%s/%s", ghReleasesURL, latest, assetName)

	tmp, err := downloadBinary(url)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("lumina: não foi possível determinar caminho do binário: %w", err)
	}

	if err := replaceSelf(self, tmp); err != nil {
		return err
	}

	fmt.Fprintf(out, "lumina atualizado para %s em %s\n", latest, self)
	return nil
}

func fetchLatestTag() (string, error) {
	resp, err := http.Get(ghAPIBase + "/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API retornou status %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", fmt.Errorf("resposta inesperada da API: %w", err)
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("nenhuma release encontrada em %s", luminaRepo)
	}
	return rel.TagName, nil
}

func downloadBinary(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("lumina: erro ao baixar binário: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lumina: asset não encontrado (status %d)\n  URL: %s", resp.StatusCode, url)
	}

	f, err := os.CreateTemp("", "lumina-update-*")
	if err != nil {
		return "", fmt.Errorf("lumina: erro ao criar arquivo temporário: %w", err)
	}
	name := f.Name()

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(name)
		return "", fmt.Errorf("lumina: erro ao escrever download: %w", err)
	}
	f.Close()

	if err := os.Chmod(name, 0755); err != nil {
		os.Remove(name)
		return "", fmt.Errorf("lumina: erro ao definir permissões: %w", err)
	}
	return name, nil
}

// replaceSelf faz rename atômico: self → self.old, newBin → self, remove .old.
// Em caso de erro no segundo rename, faz rollback restaurando self.old.
func replaceSelf(self, newBin string) error {
	backup := self + ".old"

	if err := os.Rename(self, backup); err != nil {
		return fmt.Errorf("lumina: não foi possível fazer backup do binário atual: %w", err)
	}

	if err := os.Rename(newBin, self); err != nil {
		_ = os.Rename(backup, self) // rollback
		return fmt.Errorf("lumina: não foi possível instalar nova versão: %w", err)
	}

	_ = os.Remove(backup)
	return nil
}
