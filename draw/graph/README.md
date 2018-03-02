# package graph

`import "github.com/vivint/rothko/draw/graph"`

package graph provides implementations of drawing rothko graphs.

## Usage

#### type DrawOptions

```go
type DrawOptions struct {
	// Canvas is where the drawing happens. It is expected to be large enough
	// to handle the drawing. If the canvas is nil or too small, one is
	// allocated.
	Canvas *draw.RGB

	// Columns is the set of columns to draw on the graph.
	Columns []draw.Column

	// Colors used for the heatmap.
	Colors []draw.Color
}
```

DrawOptions are all the ways you can configure the graph.

#### type MeasureOptions

```go
type MeasureOptions struct {
	// Earliest is the distribution for the earliest (closest to Now) column.
	Earliest dist.Dist

	// What time the far right of the graph represents.
	Now int64

	// Duration is the amount of history the graph represents from now.
	Duration time.Duration

	// The width of the graph
	Width int

	// The height of the graph
	Height int

	// Padding around the graph
	Padding int
}
```

MeasureOptions are options for the graph to be measured.

#### type Measured

```go
type Measured struct {
	// Bottom measured axis.
	Bottom axis.Measured

	// Observation measured axis.
	Obs observations.Measured

	// Right measured axis. Only valid if Earliest was passed with the
	// MeasureOptions.
	Right axis.Measured

	// Left measured axis.
	Left axis.Measured

	// Width, Height of the heatmap
	Width, Height int

	// The X,Y coordinates of the top left corner of the heatmap.
	X, Y int
}
```

Measured is a measured graph ready to be drawn when given data.

#### func  Measure

```go
func Measure(ctx context.Context, opts MeasureOptions) (Measured, bool)
```
Measure determines the sizes of the graph for the given parameters.

#### func (Measured) Draw

```go
func (m Measured) Draw(ctx context.Context, opts DrawOptions) *draw.RGB
```
