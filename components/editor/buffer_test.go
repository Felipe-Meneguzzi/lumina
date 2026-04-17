package editor_test

import (
	"testing"

	"github.com/menegas/lumina/components/editor"
)

func TestInsertAt_InsertsCharacter(t *testing.T) {
	b := editor.NewBuffer([]string{"hello"})
	b.InsertAt(0, 5, 'o')
	if b.Line(0) != "helloo" {
		t.Errorf("expected 'helloo', got %q", b.Line(0))
	}
}

func TestDeleteAt_RemovesCharacter(t *testing.T) {
	b := editor.NewBuffer([]string{"hello"})
	b.DeleteAt(0, 4)
	if b.Line(0) != "hell" {
		t.Errorf("expected 'hell', got %q", b.Line(0))
	}
}

func TestSplitLine_InsertsNewLine(t *testing.T) {
	b := editor.NewBuffer([]string{"hello world"})
	b.SplitLine(0, 5)
	if b.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", b.LineCount())
	}
	if b.Line(0) != "hello" {
		t.Errorf("expected line 0 = 'hello', got %q", b.Line(0))
	}
	if b.Line(1) != " world" {
		t.Errorf("expected line 1 = ' world', got %q", b.Line(1))
	}
}

func TestMoveCursor_StaysInBounds(t *testing.T) {
	b := editor.NewBuffer([]string{"hi", "there"})
	b.MoveCursor(-10, 0) // should clamp to row 0
	row, col := b.Cursor()
	if row != 0 {
		t.Errorf("expected row 0, got %d", row)
	}
	b.MoveCursor(0, 100) // should clamp to end of line
	_, col = b.Cursor()
	if col > len([]rune(b.Line(0))) {
		t.Errorf("col %d exceeds line length", col)
	}
}

func TestJoinLines_MergesLines(t *testing.T) {
	b := editor.NewBuffer([]string{"foo", "bar"})
	b.JoinLines(0)
	if b.LineCount() != 1 {
		t.Errorf("expected 1 line, got %d", b.LineCount())
	}
	if b.Line(0) != "foobar" {
		t.Errorf("expected 'foobar', got %q", b.Line(0))
	}
}
