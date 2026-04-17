package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/menegas/lumina/msgs"
)

// renderNode recursively renders a PaneNode into a string.
// Each leaf model owns its own border and dimensions (set via resize messages),
// so renderNode only handles the spatial composition of the tree.
func renderNode(n PaneNode, w, h int) string {
	if w <= 0 || h <= 0 {
		return blank(w, h)
	}

	switch v := n.(type) {
	case *LeafNode:
		return v.Model.View()

	case *SplitNode:
		ratio := clampRatio(v.Ratio)

		switch v.Direction {
		case msgs.SplitHorizontal:
			firstW := max(1, int(float64(w)*ratio))
			secondW := max(1, w-firstW)
			left := renderNode(v.First, firstW, h)
			right := renderNode(v.Second, secondW, h)
			return lipgloss.JoinHorizontal(lipgloss.Top, left, right)

		case msgs.SplitVertical:
			firstH := max(1, int(float64(h)*ratio))
			secondH := max(1, h-firstH)
			top := renderNode(v.First, w, firstH)
			bottom := renderNode(v.Second, w, secondH)
			return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
		}
	}

	return blank(w, h)
}

// blank returns a placeholder of w×h spaces used as a fallback.
func blank(w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	row := strings.Repeat(" ", w)
	rows := make([]string, h)
	for i := range rows {
		rows[i] = row
	}
	return strings.Join(rows, "\n")
}
