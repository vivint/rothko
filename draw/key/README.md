# package key

`import "github.com/vivint/rothko/draw/key"`

package key provides a way to draw a heatmap key

## Usage

#### func  Draw

```go
func Draw(canvas *draw.RGB, opts Options) (out *draw.RGB)
```
Draw draws a key using values from Options on to the provided canvas, allocating
an output canvas if the input is not large enough.

#### type Options

```go
type Options struct {
	// Colors is the slice of colors to map the column data on to.
	Colors []draw.Color

	// Height is how tall the key will be.
	Height int

	// Width is how wide the key will be.
	Width int
}
```

Options are the things you can specify to control the rendering of a key.
