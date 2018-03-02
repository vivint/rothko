# package observations

`import "github.com/vivint/rothko/draw/observations"`

package observations provides drawing of an observations axis.

## Usage

#### func  Draw

```go
func Draw(ctx context.Context, cols []draw.Column, opts Options) *draw.RGB
```
Draw renders the axis and returns a canvas allocated for the appopriate size.
See Measure if you want to control where and how it is drawn.

#### type Measured

```go
type Measured struct {
	// Width is the width in pixels of the observation axis
	Width int

	// Height is the height in pixels of the observation axis
	Height int
}
```

Measured represents a measured observations axis.

#### func  Measure

```go
func Measure(ctx context.Context, opts Options) Measured
```
Measure measures the axis sizes, and returns some state that can be used to draw
on to some canvas.

#### func (Measured) Draw

```go
func (m Measured) Draw(ctx context.Context, cols []draw.Column,
	canvas *draw.RGB) *draw.RGB
```
Draw performs the drawing of the data on to the canvas. The canvas is expected
to be large enough to handle the drawing. If the canvas is nil, one is
allocated. In either case, the canvas is returned.

#### type Options

```go
type Options struct {
	// Face is the font face to use for rendering the max observations number.
	Face font.Face

	// Width is how long the axis is.
	Width int

	// Height is the height of the bar
	Height int
}
```

Options describe the axis rendering options.
