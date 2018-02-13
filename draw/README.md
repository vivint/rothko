# draw
--
    import "github.com/spacemonkeygo/rothko/draw"

package draw provides common types for the drawing packages.

## Usage

#### type Canvas

```go
type Canvas interface {
	Set(x, y int, c Color)
	Size() (w, h int)
}
```

Canvas is the type of things that can be drawn onto.

#### type Color

```go
type Color struct {
	R, G, B uint8
}
```

Color is a simple 8 bits per channel color.

#### type Column

```go
type Column struct {
	X, W int
	Data []float64
}
```

Column represents a column to draw in a context. Data is expected to be sorted,
non-empty, and contain typical floats (no NaNs/denormals/Inf/etc).

#### type RGB

```go
type RGB struct {
	Pix    []uint8
	Stride int
	Width  int
	Height int
	X, Y   int
}
```

RGB is a byte compatabile implementation of image.RGBA, except with much less
supporting code, and no alpha channel.

#### func  NewRGB

```go
func NewRGB(w, h int) *RGB
```
NewRGB contstructs an RGB with space for the width and height.

#### func (*RGB) Raw

```go
func (m *RGB) Raw(x, y int) []uint8
```
Raw returns the raw values at the pixel, including alpha channel. It can be
mutated.

#### func (*RGB) Set

```go
func (m *RGB) Set(x, y int, c Color)
```
Set stores the pixel values in the color to the coordinate at x and y. The top
left corner is (0, 0).

#### func (*RGB) Size

```go
func (m *RGB) Size() (w, h int)
```
Size returns the width and height of the RGB.

#### func (RGB) View

```go
func (m RGB) View(x, y, w, h int) *RGB
```
View returns a view into the RGB.
