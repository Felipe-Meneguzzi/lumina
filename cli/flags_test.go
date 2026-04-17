package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Felipe-Meneguzzi/lumina/cli"
)

func TestEffectiveMaxPanes_DefaultsTo4(t *testing.T) {
	o := cli.StartupOverrides{}
	if got := o.EffectiveMaxPanes(); got != 4 {
		t.Errorf("expected 4, got %d", got)
	}
}

func TestEffectiveMaxPanes_ExplicitMP(t *testing.T) {
	o := cli.StartupOverrides{MaxPanes: 10}
	if got := o.EffectiveMaxPanes(); got != 10 {
		t.Errorf("expected 10, got %d", got)
	}
}

func TestEffectiveMaxPanes_AutoRaiseFromSP(t *testing.T) {
	o := cli.StartupOverrides{StartPanes: 5}
	if got := o.EffectiveMaxPanes(); got != 5 {
		t.Errorf("expected 5 (auto-raise), got %d", got)
	}
}

func TestEffectiveMaxPanes_MPWinsOverSP(t *testing.T) {
	o := cli.StartupOverrides{MaxPanes: 3, StartPanes: 2}
	if got := o.EffectiveMaxPanes(); got != 3 {
		t.Errorf("expected 3 (explicit -mp wins), got %d", got)
	}
}

func TestParseArgs_MPFlag_Valid(t *testing.T) {
	var buf bytes.Buffer
	o, err := cli.ParseArgs([]string{"-mp", "10"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.MaxPanes != 10 {
		t.Errorf("expected MaxPanes=10, got %d", o.MaxPanes)
	}
}

func TestParseArgs_MPFlag_Invalid(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string // substring expected in error (case-insensitive)
	}{
		{"zero", []string{"-mp", "0"}, `-mp inválido`},
		{"negative", []string{"-mp", "-1"}, `-mp inválido`},
		{"non-numeric", []string{"-mp", "abc"}, `invalid`}, // flag pkg rejects before our check
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := cli.ParseArgs(tc.args, &buf)
			if err == nil {
				t.Fatalf("expected error for args %v, got nil", tc.args)
			}
			combined := err.Error() + buf.String()
			if !strings.Contains(strings.ToLower(combined), strings.ToLower(tc.want)) {
				t.Errorf("expected error mentioning %q, got %q", tc.want, combined)
			}
		})
	}
}

func TestParseStartPanes_Valid(t *testing.T) {
	cases := []struct {
		in     string
		orient cli.Orient
		count  int
	}{
		{"h1", cli.OrientHorizontal, 1},
		{"h3", cli.OrientHorizontal, 3},
		{"v2", cli.OrientVertical, 2},
		{"v99", cli.OrientVertical, 99},
		{"h10", cli.OrientHorizontal, 10},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			o, n, err := cli.ParseStartPanes(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if o != tc.orient || n != tc.count {
				t.Errorf("expected (%v, %d), got (%v, %d)", tc.orient, tc.count, o, n)
			}
		})
	}
}

func TestParseStartPanes_Invalid(t *testing.T) {
	invalid := []string{"", "h", "v", "3", "d2", "h0", "h-1", "hABC", "horizontal"}
	for _, in := range invalid {
		t.Run(in, func(t *testing.T) {
			_, _, err := cli.ParseStartPanes(in)
			if err == nil {
				t.Errorf("expected error for %q, got nil", in)
			}
		})
	}
}

func TestValidate_Conflict_MPBelowSP(t *testing.T) {
	o := cli.StartupOverrides{MaxPanes: 2, StartPanes: 5, StartOrient: cli.OrientHorizontal}
	err := o.Validate()
	if err == nil {
		t.Fatal("expected error for -mp 2 + -sp h5")
	}
	if !strings.Contains(err.Error(), "excede") {
		t.Errorf("expected error mentioning 'excede', got %q", err.Error())
	}
}

func TestValidate_NoConflict(t *testing.T) {
	cases := []cli.StartupOverrides{
		{},
		{MaxPanes: 10},
		{StartPanes: 3, StartOrient: cli.OrientHorizontal},
		{MaxPanes: 10, StartPanes: 3, StartOrient: cli.OrientHorizontal},
	}
	for _, o := range cases {
		if err := o.Validate(); err != nil {
			t.Errorf("unexpected error for %+v: %v", o, err)
		}
	}
}

func TestParseArgs_SPFlag(t *testing.T) {
	var buf bytes.Buffer
	o, err := cli.ParseArgs([]string{"-sp", "h3"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.StartPanes != 3 || o.StartOrient != cli.OrientHorizontal {
		t.Errorf("expected h3 parsed as (Horizontal,3), got (%v,%d)", o.StartOrient, o.StartPanes)
	}
}

func TestParseArgs_SPFlag_Invalid(t *testing.T) {
	var buf bytes.Buffer
	_, err := cli.ParseArgs([]string{"-sp", "d3"}, &buf)
	if err == nil {
		t.Fatal("expected error for -sp d3")
	}
}

func TestParseArgs_Conflict_MP_SP(t *testing.T) {
	var buf bytes.Buffer
	_, err := cli.ParseArgs([]string{"-mp", "2", "-sp", "h5"}, &buf)
	if err == nil {
		t.Fatal("expected error for -mp 2 -sp h5")
	}
	if !strings.Contains(err.Error(), "excede") {
		t.Errorf("expected 'excede' in error, got %q", err.Error())
	}
}

func TestParseArgs_SCFlag_Valid(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"simple", []string{"-sc", "claude"}, "claude"},
		{"with args", []string{"-sc", "claude --model opus"}, "claude --model opus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			o, err := cli.ParseArgs(tc.args, &buf)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if o.StartCommand != tc.want {
				t.Errorf("expected %q, got %q", tc.want, o.StartCommand)
			}
		})
	}
}

func TestParseArgs_SCFlag_Empty(t *testing.T) {
	var buf bytes.Buffer
	_, err := cli.ParseArgs([]string{"-sc", ""}, &buf)
	if err == nil {
		t.Fatal("expected error for empty -sc")
	}
	if !strings.Contains(err.Error(), "vazio") {
		t.Errorf("expected 'vazio' in error, got %q", err.Error())
	}
}

func TestParseArgs_FilePositional(t *testing.T) {
	var buf bytes.Buffer
	o, err := cli.ParseArgs([]string{"-sp", "h2", "notes.md"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.FilePath != "notes.md" {
		t.Errorf("expected FilePath=notes.md, got %q", o.FilePath)
	}
	if o.StartPanes != 2 {
		t.Errorf("expected StartPanes=2, got %d", o.StartPanes)
	}
}

func TestParseArgs_Full_Example(t *testing.T) {
	var buf bytes.Buffer
	o, err := cli.ParseArgs([]string{"-mp", "10", "-sp", "h3", "-sc", "claude"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.MaxPanes != 10 || o.StartPanes != 3 || o.StartOrient != cli.OrientHorizontal || o.StartCommand != "claude" {
		t.Errorf("unexpected overrides: %+v", o)
	}
	if got := o.EffectiveMaxPanes(); got != 10 {
		t.Errorf("EffectiveMaxPanes: expected 10, got %d", got)
	}
}
