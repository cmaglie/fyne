package layout

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*weightedGridLayout)(nil)

type weightedGridLayout struct {
	Cols            []int
	TotalCols       int
	vertical, adapt bool
}

func sum(cols []int) int {
	res := 0
	for _, col := range cols {
		res += col
	}
	return res
}

// NewAdaptiveWeightedGridLayout returns a new grid layout which uses columns when horizontal but rows when vertical.
func NewAdaptiveWeightedGridLayout(rowcols []int) fyne.Layout {
	return &weightedGridLayout{Cols: rowcols, TotalCols: sum(rowcols), adapt: true}
}

// NewWeightedGridLayout returns a grid layout arranged in a specified number of columns.
// The number of rows will depend on how many children are in the container that uses this layout.
func NewWeightedGridLayout(cols []int) fyne.Layout {
	return NewWeightedGridLayoutWithColumns(cols)
}

// NewWeightedGridLayoutWithColumns returns a new grid layout that specifies a column count and wrap to new rows when needed.
func NewWeightedGridLayoutWithColumns(cols []int) fyne.Layout {
	return &weightedGridLayout{Cols: cols, TotalCols: sum(cols)}
}

// NewWeightedGridLayoutWithRows returns a new grid layout that specifies a row count that creates new rows as required.
func NewWeightedGridLayoutWithRows(rows []int) fyne.Layout {
	return &weightedGridLayout{Cols: rows, TotalCols: sum(rows), vertical: true}
}

func (g *weightedGridLayout) horizontal() bool {
	if g.adapt {
		return fyne.IsHorizontal(fyne.CurrentDevice().Orientation())
	}

	return !g.vertical
}

func (g *weightedGridLayout) countRows(objects []fyne.CanvasObject) int {
	count := 0
	for i, child := range objects {
		if child.Visible() {
			count += g.Cols[i]
		}
	}

	return int(math.Ceil(float64(count) / float64(g.TotalCols)))
}

// Layout is called to pack all child objects into a specified size.
// For a GridLayout this will pack objects into a table format with the number
// of columns specified in our constructor.
func (g *weightedGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	rows := g.countRows(objects)
	fmt.Println("SIZE", size)
	padWidth := float32(g.TotalCols-1) * theme.Padding()
	padHeight := float32(rows-1) * theme.Padding()
	cellWidth := float64(size.Width-padWidth) / float64(g.TotalCols)
	cellHeight := float64(size.Height-padHeight) / float64(rows)

	if !g.horizontal() {
		padWidth, padHeight = padHeight, padWidth
		cellWidth = float64(size.Width-padWidth) / float64(rows)
		cellHeight = float64(size.Height-padHeight) / float64(g.TotalCols)
	}

	row, col := 0, 0
	i := 0
	for idx, child := range objects {
		if !child.Visible() {
			continue
		}

		span := g.Cols[idx]
		colSpan, rowSpan := 1, 1
		if g.horizontal() {
			colSpan = span
		} else {
			rowSpan = span
		}
		x1 := getLeading(cellWidth, col)
		y1 := getLeading(cellHeight, row)
		x2 := getTrailing(cellWidth, col+colSpan-1)
		y2 := getTrailing(cellHeight, row+rowSpan-1)
		fmt.Println(x1, y1, "-", x2, y2)
		child.Move(fyne.NewPos(x1, y1))
		child.Resize(fyne.NewSize(x2-x1, y2-y1))

		for span > 0 {
			if g.horizontal() {
				if (i+1)%g.TotalCols == 0 {
					row++
					col = 0
				} else {
					col++
				}
			} else {
				if (i+1)%g.TotalCols == 0 {
					col++
					row = 0
				} else {
					row++
				}
			}
			i++
			span--
		}
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For a GridLayout this is the size of the largest child object multiplied by
// the required number of columns and rows, with appropriate padding between
// children.
func (g *weightedGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	rows := g.countRows(objects)
	minSize := fyne.NewSize(0, 0)
	for i, child := range objects {
		if !child.Visible() {
			continue
		}
		childMinSize := child.MinSize()
		childMinSize.Height /= float32(g.Cols[i])
		minSize = minSize.Max(childMinSize)
	}
	if g.horizontal() {
		minContentSize := fyne.NewSize(minSize.Width*float32(g.TotalCols), minSize.Height*float32(rows))
		return minContentSize.Add(fyne.NewSize(theme.Padding()*fyne.Max(float32(g.TotalCols-1), 0), theme.Padding()*fyne.Max(float32(rows-1), 0)))
	}

	minContentSize := fyne.NewSize(minSize.Width*float32(rows), minSize.Height*float32(g.TotalCols))
	return minContentSize.Add(fyne.NewSize(theme.Padding()*fyne.Max(float32(rows-1), 0), theme.Padding()*fyne.Max(float32(g.TotalCols-1), 0)))
}
