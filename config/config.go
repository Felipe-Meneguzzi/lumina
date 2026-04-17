package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds user-configurable settings for Lumina.
type Config struct {
	Shell           string      `toml:"shell"`
	MetricsInterval int         `toml:"metrics_interval"` // milliseconds
	ShowHidden      bool        `toml:"show_hidden"`
	SidebarWidth    int         `toml:"sidebar_width"` // columns
	Theme           string      `toml:"theme"`
	ForceShellTheme bool        `toml:"force_shell_theme"` // inject Lumina's default prompt into spawned shells
	Keys            Keybindings `toml:"-"`                 // loaded separately from keybindings.json
	ShellWarning    string      `toml:"-"`                 // set when configured shell was rejected
}

// isWindowsExecutable reports whether a path looks like a Windows PE binary.
// On Linux/WSL, Windows binaries (.exe) are accessible via binfmt_misc but cannot
// host a proper POSIX PTY session — they open their own console window instead.
func isWindowsExecutable(s string) bool {
	return runtime.GOOS != "windows" && strings.HasSuffix(strings.ToLower(s), ".exe")
}

// isWSL reports whether the process is running inside Windows Subsystem for Linux.
func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// validateShell returns the first usable POSIX shell from the candidate list.
// Candidates: configured value → $SHELL env → /bin/bash → /bin/zsh → /bin/sh.
// On Linux/WSL, Windows executables (.exe) are always rejected even if found in PATH.
func validateShell(configured string) string {
	candidates := []string{configured, os.Getenv("SHELL"), "/bin/bash", "/bin/zsh", "/bin/sh"}
	for _, s := range candidates {
		if s == "" {
			continue
		}
		// Reject Windows PE binaries on Linux/WSL — they can't provide a PTY session.
		if isWindowsExecutable(s) {
			continue
		}
		if _, err := exec.LookPath(s); err == nil {
			return s
		}
	}
	return "/bin/sh" // always present on POSIX systems
}

func defaults() Config {
	return Config{
		Shell:           validateShell(""),
		MetricsInterval: 1000,
		ShowHidden:      true,
		SidebarWidth:    30,
		Theme:           "default",
		ForceShellTheme: true,
	}
}

// LoadConfig reads ~/.config/lumina/config.toml and keybindings.json,
// falling back to built-in defaults for any missing values.
func LoadConfig() (Config, error) {
	cfg := defaults()

	home, err := os.UserHomeDir()
	if err != nil {
		kb, _ := LoadKeybindings()
		cfg.Keys = kb
		return cfg, nil //nolint:nilerr
	}

	path := filepath.Join(home, ".config", "lumina", "config.toml")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return cfg, err
		}
		requested := cfg.Shell
		cfg.Shell = validateShell(cfg.Shell) // validate after user config override
		// Surface a diagnostic when the configured shell was silently replaced.
		if requested != "" && requested != cfg.Shell && isWindowsExecutable(requested) {
			// Store the warning so app.go can display it in the status bar on startup.
			cfg.ShellWarning = "WSL: shell '" + requested + "' é um binário Windows — usando " + cfg.Shell
		}
	}

	kb, err := LoadKeybindings()
	if err != nil {
		return cfg, err
	}
	cfg.Keys = kb
	return cfg, nil
}
