// Package cli parses Lumina startup flags (-mp, -sp, -sc, -nsb) into a
// StartupOverrides value consumed by main.
package cli

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Orient is the orientation requested by -sp.
type Orient int

const (
	OrientNone Orient = iota
	OrientHorizontal
	OrientVertical
)

const defaultMaxPanes = 4

// StartupOverrides captures the CLI-provided session overrides.
type StartupOverrides struct {
	MaxPanes     int    // 0 = use default (4); >0 = explicit override.
	StartPanes   int    // 0 or 1 = no pre-split; >=2 = create N initial panes.
	StartOrient  Orient // OrientNone when StartPanes <= 1.
	StartCommand string // "" = shell default; otherwise run this command in initial panes.
	NoSidebar    bool   // true = hide sidebar on startup for all panes.
	FilePath     string // positional argument (compatibility with `lumina <file>`).
}

// EffectiveMaxPanes returns the max-panes ceiling to apply to the session:
// explicit -mp wins, otherwise default 4 auto-raised to fit -sp.
func (o StartupOverrides) EffectiveMaxPanes() int {
	if o.MaxPanes > 0 {
		return o.MaxPanes
	}
	if o.StartPanes > defaultMaxPanes {
		return o.StartPanes
	}
	return defaultMaxPanes
}

// Validate enforces cross-flag invariants (e.g. -sp count within -mp ceiling).
func (o StartupOverrides) Validate() error {
	if o.MaxPanes > 0 && o.StartPanes > o.MaxPanes {
		return fmt.Errorf(
			"lumina: -sp %s%d excede -mp %d: não é possível criar %d painéis iniciais com teto %d",
			orientLetter(o.StartOrient), o.StartPanes, o.MaxPanes, o.StartPanes, o.MaxPanes,
		)
	}
	return nil
}

func orientLetter(o Orient) string {
	if o == OrientVertical {
		return "v"
	}
	return "h"
}

// ParseStartPanes parses the -sp value (e.g. "h3", "v2") into its components.
func ParseStartPanes(s string) (Orient, int, error) {
	if len(s) < 2 {
		return OrientNone, 0, newSPError(s)
	}
	var orient Orient
	switch s[0] {
	case 'h':
		orient = OrientHorizontal
	case 'v':
		orient = OrientVertical
	default:
		return OrientNone, 0, newSPError(s)
	}
	n, err := strconv.Atoi(s[1:])
	if err != nil || n < 1 {
		return OrientNone, 0, newSPError(s)
	}
	return orient, n, nil
}

func newSPError(got string) error {
	return fmt.Errorf(
		`lumina: -sp inválido: esperado h<N> ou v<N> com N >= 1, recebi %q`, got,
	)
}

// ParseArgs parses args (typically os.Args[1:]) into StartupOverrides.
// Writes usage/errors to errOut when flag parsing itself fails.
// Returns a descriptive error (without prefix "lumina: " duplicated) on
// validation problems so main can print it to stderr and exit(2).
func ParseArgs(args []string, errOut io.Writer) (StartupOverrides, error) {
	fs := flag.NewFlagSet("lumina", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.Usage = func() { fmt.Fprint(errOut, usageText()) }

	var (
		mp    int
		spRaw string
		sc    string
		nsb   bool
	)
	fs.IntVar(&mp, "mp", 0, "max panes allowed in this session (default 4)")
	fs.StringVar(&spRaw, "sp", "", "pre-split layout: h<N> (horizontal) or v<N> (vertical)")
	fs.StringVar(&sc, "sc", "", "run <command> in initial panes instead of the default shell")
	fs.BoolVar(&nsb, "nsb", false, "hide sidebar on startup")

	if err := fs.Parse(args); err != nil {
		return StartupOverrides{}, err
	}

	setFlags := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })

	out := StartupOverrides{}

	if setFlags["mp"] {
		if mp < 1 {
			return out, fmt.Errorf(
				`lumina: -mp inválido: esperado inteiro >= 1, recebi %q`, strconv.Itoa(mp),
			)
		}
		out.MaxPanes = mp
	}

	if spRaw != "" {
		orient, count, err := ParseStartPanes(spRaw)
		if err != nil {
			return out, err
		}
		out.StartOrient = orient
		out.StartPanes = count
	}

	if setFlags["sc"] {
		if sc == "" {
			return out, fmt.Errorf("lumina: -sc inválido: comando não pode ser vazio")
		}
		out.StartCommand = strings.TrimSpace(sc)
	}

	if nsb {
		out.NoSidebar = true
	}

	if rest := fs.Args(); len(rest) > 0 {
		out.FilePath = rest[0]
	}

	if err := out.Validate(); err != nil {
		return out, err
	}
	return out, nil
}

func usageText() string {
	return `Lumina — TUI editor with splittable panes.

Usage:
  lumina [flags] [file]

Flags:
  -mp N             Max panes allowed in this session (default: 4)
  -sp <h|v>N        Pre-split layout on startup (e.g. h3 = 3 horizontal panes)
  -sc <command>     Run <command> in initial panes instead of the default shell
                    (applies only to panes created by -sp; later splits use the shell)
  -nsb              Hide sidebar on startup for all panes
  --version, -v     Print version and exit
  --update          Check for updates and install if a newer release is available
  --help, -h        Show this help

Examples:
  lumina
  lumina -mp 10 -sp h3 -sc claude
  lumina notes.md -sp v2
  lumina -nsb -sp h2
`
}

// UsageText exposes the help block for main's custom --help/-h handler.
func UsageText() string { return usageText() }
