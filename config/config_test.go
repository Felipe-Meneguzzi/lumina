package config

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestValidateShell(t *testing.T) {
	t.Run("valid shell is returned as-is", func(t *testing.T) {
		got := validateShell("/bin/sh")
		if got != "/bin/sh" {
			t.Errorf("expected /bin/sh, got %q", got)
		}
	})

	t.Run("invalid shell falls back to a valid one", func(t *testing.T) {
		got := validateShell("invalid-shell-xyz-not-a-real-binary")
		if got == "" || got == "invalid-shell-xyz-not-a-real-binary" {
			t.Errorf("expected a valid fallback shell, got %q", got)
		}
	})

	t.Run("empty configured shell uses SHELL env var", func(t *testing.T) {
		orig := os.Getenv("SHELL")
		_ = os.Setenv("SHELL", "/bin/sh")
		defer func() { _ = os.Setenv("SHELL", orig) }()

		got := validateShell("")
		if got == "" {
			t.Error("expected a valid shell, got empty string")
		}
	})

	t.Run("empty configured shell and empty SHELL env falls back to bin/sh", func(t *testing.T) {
		orig := os.Getenv("SHELL")
		_ = os.Unsetenv("SHELL")
		defer func() { _ = os.Setenv("SHELL", orig) }()

		got := validateShell("")
		if got == "" {
			t.Error("expected fallback shell, got empty string")
		}
	})

	t.Run("result is always an executable that exists", func(t *testing.T) {
		got := validateShell("totally-nonexistent-shell")
		if _, err := os.Stat(got); err != nil {
			t.Errorf("returned shell %q does not exist: %v", got, err)
		}
	})

	t.Run("windows .exe is rejected on linux/wsl", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("only meaningful on Linux/WSL")
		}
		got := validateShell("powershell.exe")
		if strings.HasSuffix(strings.ToLower(got), ".exe") {
			t.Errorf("expected non-.exe fallback on Linux, got %q", got)
		}
	})

	t.Run("cmd.exe is rejected on linux/wsl", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("only meaningful on Linux/WSL")
		}
		got := validateShell("cmd.exe")
		if strings.HasSuffix(strings.ToLower(got), ".exe") {
			t.Errorf("expected non-.exe fallback on Linux, got %q", got)
		}
	})
}
