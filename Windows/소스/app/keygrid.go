package main

// Logical key order is row-major for both models. PN15 physical key centers
// descend as each column moves to the right (5.93 mm Y shift / 19 mm pitch).
// The visual drop is independent of the 3x5 electrical switch matrix.
type keyGrid struct{ Columns, Rows, StartX, StartY, CellW, CellH, GapX, GapY, ColumnDrop int }

func gridForKeys(keys int) keyGrid {
	if keys == 15 {
		return keyGrid{Columns: 5, Rows: 3, StartX: 308, StartY: 297,
			CellW: 100, CellH: 100, GapX: 10, GapY: 8, ColumnDrop: 32}
	}
	return keyGrid{Columns: 3, Rows: 4, StartX: 393, StartY: 292,
		CellW: 114, CellH: 114, GapX: 13, GapY: 4}
}
func (g keyGrid) pos(idx int) (int, int) {
	col := idx % g.Columns
	row := idx / g.Columns
	return g.StartX + col*(g.CellW+g.GapX), g.StartY + row*(g.CellH+g.GapY) + col*g.ColumnDrop
}

// A mirrored physical housing keeps the same row-major electrical key IDs.
// Only visual Y-offset changes with firmware-advertised handedness.
func gridForHand(keys int, side string) keyGrid { return gridForKeys(keys) }
func (g keyGrid) posForSide(idx int, side string) (int, int) {
	x, y := g.pos(idx)
	if g.Columns == 5 && validSide(side) == "L" {
		col := idx % g.Columns
		y += (g.Columns - 1 - 2*col) * g.ColumnDrop
	}
	return x, y
}
