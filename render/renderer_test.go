package render

import (
	"testing"

	"github.com/j-p/recliner/vdom"
)

func TestBoxMargins(t *testing.T) {
	r := NewRenderer()
	r.SetSize(20, 10)

	node := &vdom.Element{
		Type: "box",
		Style: vdom.Style{
			Width:        10,
			Height:       5,
			MarginLeft:   2,
			MarginTop:    1,
			Background:   "red",
			BorderTop:    true,
			BorderBottom: true,
			BorderLeft:   true,
			BorderRight:  true,
		},
		Props: map[string]any{
			"borderStyle": "single",
		},
	}

	r.Render(node)

	buf := r.lastBuffer
	if buf == nil {
		t.Fatal("Buffer is nil after render")
	}

	// Check that the box starts at (2, 1) due to margins
	cell := buf.Cells[1][2]
	if cell.Char != '┌' {
		t.Errorf("Expected '┌' at (2, 1), got %c", cell.Char)
	}

	// Content should start at (3, 2) because of border (bs=1, p=0)
	cell = buf.Cells[2][3]
	if cell.Style.Background != "red" {
		t.Errorf("Expected red background at (3, 2), got %s", cell.Style.Background)
	}

	// Margin area should NOT have the box background
	cell = buf.Cells[1][0]
	if cell.Style.Background == "red" {
		t.Errorf("Expected NO red background at (0, 1) (margin area)")
	}
}

func TestBoxMarginsOverlap(t *testing.T) {
	r := NewRenderer()
	r.SetSize(40, 10)

	node := &vdom.Element{
		Type: "box",
		Style: vdom.Style{
			Display:       "flex",
			FlexDirection: "row",
			Width:         40,
			Height:        10,
		},
		Children: []vdom.Node{
			&vdom.Element{
				Type: "box",
				Style: vdom.Style{
					Width:        10,
					Height:       5,
					MarginLeft:   2,
					Background:   "red",
					BorderTop:    true,
					BorderBottom: true,
					BorderLeft:   true,
					BorderRight:  true,
				},
				Props: map[string]any{"borderStyle": "single"},
			},
			&vdom.Element{
				Type: "box",
				Style: vdom.Style{
					Width:        10,
					Height:       5,
					Background:   "blue",
					BorderTop:    true,
					BorderBottom: true,
					BorderLeft:   true,
					BorderRight:  true,
				},
				Props: map[string]any{"borderStyle": "single"},
			},
		},
	}

	r.Render(node)
	buf := r.lastBuffer

	// Box A starts at (2, 0) relative to parent content.
	// Parent here has no border/padding by default since we didn't specify.
	// Wait, root box has no border by default if not specified?
	// No, measureBox defaults to bs=1 if not specified and not "none".
	// So parent has bs=1. Content starts at (1, 1).

	// Box A total width including margin is 12.
	// Box B total width is 10.

	// Box A box area starts at (1 + MarginLeft, 1 + MarginTop) = (1+2, 1+0) = (3, 1).
	cellA := buf.Cells[1][3]
	if cellA.Char != '┌' {
		t.Errorf("Box A should start at (3, 1), got %c", cellA.Char)
	}

	// Box B box area starts at (1 + TotalWidthA, 1 + MarginTop) = (1+12, 1+0) = (13, 1).
	cellB := buf.Cells[1][13]
	if cellB.Char != '┌' {
		t.Errorf("Box B should start at (13, 1), got %c", cellB.Char)
	}
}
