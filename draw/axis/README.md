# package axis

`import "github.com/spacemonkeygo/rothko/draw/axis"`

package axis provides implementations of drawing rothko axes.

## Usage

#### func  Draw

```go
func Draw(ctx context.Context, opts Options) *draw.RGB
```
Draw renders the axis and returns a canvas allocated for the appopriate size.
See Measure if you want to control where and how it is drawn.

#### type Label

```go
type Label struct {
	// Position is the position of the tick mark as a float in [0, 1].
	Position float64

	// Text is the text of the tick mark.
	Text string
}
```

Label represents a tick mark on the axis.

#### type Measured

```go
type Measured struct {
	// Width is the width in pixels of the drawn axis.
	Width int

	// Height is the height in pixels of the drawn axis
	Height int
}
```


#### func  Measure

```go
func Measure(ctx context.Context, opts Options) Measured
```
Measure measures the axis sizes, and returns some state that can be used to draw
on to some canvas.

#### func (Measured) Draw

```go
func (m Measured) Draw(ctx context.Context, canvas *draw.RGB) *draw.RGB
```
Draw performs the drawing of the data on to the canvas. The canvas is expected
to be large enough to handle the drawing. If the canvas is nil, one is
allocated. In either case, the canvas is returned.

#### type Options

```go
type Options struct {
	// Face is the font face to use for rendering the label text.
	Face font.Face

	// Labels is the set of labels to draw.
	Labels []Label

	// Vertical is if the axis is vertical.
	Vertical bool

	// Length is how long the axis is.
	Length int

	// If true, vertical axes will be drawn for the left size. Horizontal axes
	// ignore this field.
	Flip bool

	// If true, the label text will not go past the boundaries of Length.
	DontBleed bool
}
```

Options describe the axis rendering options.
