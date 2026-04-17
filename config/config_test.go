package config

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
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

// withTempHome temporarily redirects HOME so LoadConfig's os.UserHomeDir()
// resolves into an isolated test directory. Returns the path of that directory.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

// TestEditor_Default verifies FR-018 path (a): when config.toml does not set
// editor, LoadConfig fills in the default "nano".
func TestEditor_Default(t *testing.T) {
	withTempHome(t)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Editor != "nano" {
		t.Errorf("expected Editor=\"nano\" by default, got %q", cfg.Editor)
	}
}

// TestEditor_EmptyTreatedAsAbsent verifies FR-018 path (b): an explicit empty
// string in config.toml is treated as "absent" and falls back to "nano".
func TestEditor_EmptyTreatedAsAbsent(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lumina")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("editor = \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Editor != "nano" {
		t.Errorf("expected empty editor to fall back to \"nano\", got %q", cfg.Editor)
	}
}

// TestEditor_ExplicitValue_RoundTrips verifies FR-018 path (c): a concrete
// value survives TOML encoding + LoadConfig unchanged.
func TestEditor_ExplicitValue_RoundTrips(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lumina")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Encode a struct containing the desired editor value so we test the real
	// TOML encoder used by writeDefaults.
	src := struct {
		Editor string `toml:"editor"`
	}{Editor: "vim"}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(src); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Editor != "vim" {
		t.Errorf("expected Editor=\"vim\" after round-trip, got %q", cfg.Editor)
	}
}

// TestEditor_UnresolvedPathNotRewritten verifies FR-018 path (d): a binary
// not present in PATH is preserved verbatim — resolution is deferred to the
// spawn site (app.openInExternalEditor), where a missing binary surfaces as
// a StatusBarNotifyMsg instead of a silent rewrite to the default.
func TestEditor_UnresolvedPathNotRewritten(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lumina")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"),
		[]byte("editor = \"/absolutely/not/real/xyz\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Editor != "/absolutely/not/real/xyz" {
		t.Errorf("expected unresolved editor to be preserved, got %q", cfg.Editor)
	}
}
