package termui

import (
	"image"
	"sort"

	drawille "github.com/cjbassi/drawille-go"
	. "github.com/gizak/termui"
)

// LineGraph implements a line graph of data points.
type LineGraph struct {
	*Block
	Data      map[string][]float64
	LineColor map[string]Attribute
	Zoom      int
	Labels    map[string]string

	DefaultLineColor Attribute
}

// NewLineGraph returns a new LineGraph with current theme.
func NewLineGraph() *LineGraph {
	return &LineGraph{
		Block:     NewBlock(),
		Data:      make(map[string][]float64),
		LineColor: make(map[string]Attribute),
		Labels:    make(map[string]string),
		Zoom:      5,
	}
}

func (self *LineGraph) Draw(buf *Buffer) {
	self.Block.Draw(buf)
	// we render each data point on to the canvas then copy over the braille to the buffer at the end
	// fyi braille characters have 2x4 dots for each character
	c := drawille.NewCanvas()
	// used to keep track of the braille colors until the end when we render the braille to the buffer
	colors := make([][]Attribute, self.Inner.Dx()+2)
	for i := range colors {
		colors[i] = make([]Attribute, self.Inner.Dy()+2)
	}

	// sort the series so that overlapping data will overlap the same way each time
	seriesList := make([]string, len(self.Data))
	i := 0
	for seriesName := range self.Data {
		seriesList[i] = seriesName
		i++
	}
	sort.Strings(seriesList)

	// draw lines in reverse order so that the first color defined in the colorscheme is on top
	for i := len(seriesList) - 1; i >= 0; i-- {
		seriesName := seriesList[i]
		seriesData := self.Data[seriesName]
		seriesLineColor, ok := self.LineColor[seriesName]
		if !ok {
			seriesLineColor = self.DefaultLineColor
		}

		// coordinates of last point
		lastY, lastX := -1, -1
		// assign colors to `colors` and lines/points to the canvas
		for i := len(seriesData) - 1; i >= 0; i-- {
			x := ((self.Inner.Dx() + 1) * 2) - 1 - (((len(seriesData) - 1) - i) * self.Zoom)
			y := ((self.Inner.Dy() + 1) * 4) - 1 - int((float64((self.Inner.Dy())*4)-1)*(seriesData[i]/100))
			if x < 0 {
				// render the line to the last point up to the wall
				if x > 0-self.Zoom {
					for _, p := range drawille.Line(lastX, lastY, x, y) {
						if p.X > 0 {
							c.Set(p.X, p.Y)
							colors[p.X/2][p.Y/4] = seriesLineColor
						}
					}
				}
				break
			}
			if lastY == -1 { // if this is the first point
				c.Set(x, y)
				colors[x/2][y/4] = seriesLineColor
			} else {
				c.DrawLine(lastX, lastY, x, y)
				for _, p := range drawille.Line(lastX, lastY, x, y) {
					colors[p.X/2][p.Y/4] = seriesLineColor
				}
			}
			lastX, lastY = x, y
		}

		// copy braille and colors to buffer
		for y, line := range c.Rows(c.MinX(), c.MinY(), c.MaxX(), c.MaxY()) {
			for x, char := range line {
				x /= 3 // idk why but it works
				if x == 0 {
					continue
				}
				if char != 10240 { // empty braille character
					buf.SetCell(
						Cell{char, AttrPair{colors[x][y], -1}},
						image.Pt(self.Inner.Min.X+x-1, self.Inner.Min.Y+y-1),
					)
				}
			}
		}
	}

	// renders key/label ontop
	for i, seriesName := range seriesList {
		if i+2 > self.Inner.Dy() {
			continue
		}
		seriesLineColor, ok := self.LineColor[seriesName]
		if !ok {
			seriesLineColor = self.DefaultLineColor
		}

		// render key ontop, but let braille be drawn over space characters
		str := seriesName + " " + self.Labels[seriesName]
		for k, char := range str {
			if char != ' ' {
				buf.SetCell(
					Cell{char, AttrPair{seriesLineColor, -1}},
					image.Pt(self.Inner.Min.X+2+k, self.Inner.Min.Y+i+1),
				)
			}
		}

	}
}
