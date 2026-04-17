package terminal

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed themes/bashrc
var luminaBashrc string

//go:embed themes/zshrc
var luminaZshrc string

// themePaths holds the paths to Lumina's shell theme files after being written to disk.
type themePaths struct {
	bashrc  string // bash --rcfile target
	zdotdir string // directory used as ZDOTDIR for zsh
}

// ensureThemePaths writes Lumina's bundled shell theme files under
// ~/.config/lumina/shell/ and returns the resolved paths. The files are
// rewritten on every call so upgrades ship new prompts transparently.
func ensureThemePaths() (themePaths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return themePaths{}, err
	}
	base := filepath.Join(home, ".config", "lumina", "shell")
	zdotdir := filepath.Join(base, "zdotdir")
	if err := os.MkdirAll(zdotdir, 0o755); err != nil {
		return themePaths{}, err
	}

	bashrc := filepath.Join(base, "bashrc")
	if err := os.WriteFile(bashrc, []byte(luminaBashrc), 0o644); err != nil {
		return themePaths{}, err
	}
	zshrc := filepath.Join(zdotdir, ".zshrc")
	if err := os.WriteFile(zshrc, []byte(luminaZshrc), 0o644); err != nil {
		return themePaths{}, err
	}
	return themePaths{bashrc: bashrc, zdotdir: zdotdir}, nil
}

// buildShellCommand returns the exec.Cmd used to spawn the interactive shell.
// When forceTheme is true and the shell is bash or zsh, it wires up Lumina's
// themed rcfile so every pane renders the same prompt regardless of the user's
// own dotfiles. For any other shell (or on theme-write failure) it falls back
// to a vanilla command so the terminal still works.
func buildShellCommand(shell string, forceTheme bool, env *[]string) *exec.Cmd {
	if !forceTheme {
		return exec.Command(shell)
	}
	paths, err := ensureThemePaths()
	if err != nil {
		return exec.Command(shell)
	}
	switch filepath.Base(shell) {
	case "bash":
		return exec.Command(shell, "--rcfile", paths.bashrc, "-i")
	case "zsh":
		*env = setEnv(*env, "ZDOTDIR", paths.zdotdir)
		return exec.Command(shell, "-i")
	default:
		return exec.Command(shell)
	}
}
