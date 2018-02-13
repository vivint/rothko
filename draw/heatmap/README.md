# package heatmap

`import "github.com/spacemonkeygo/rothko/draw/heatmap"`

package heatmap provides implementations of drawing rothko heatmaps.

## Usage

#### type Heatmap

```go
type Heatmap struct {
}
```

Heatmap is a struct that draws heatmaps from provided columns.

#### func  New

```go
func New(opts Options) *Heatmap
```
New returns a new Heatmap using the given options.

#### func (*Heatmap) Draw

```go
func (d *Heatmap) Draw(ctx context.Context, col draw.Column)
```
Draw writes the column to the canvas.

#### type Options

```go
type Options struct {
	// Colors is the slice of colors to map the column data on to.
	Colors []draw.Color

	// Canvas to draw on to
	Canvas draw.Canvas

	// Map takes a value from the Data in the column, and expects it to be
	// mapped to a value in [0,1] specifying the color.
	Map func(float64) float64
}
```

Options are the things you can specify to control the rendering of a heatmap.
