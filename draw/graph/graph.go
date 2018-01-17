// Copyright (C) 2017. See AUTHORS.

package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/data/merge"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/spacemonkeygo/rothko/draw/axis"
	"github.com/spacemonkeygo/rothko/draw/heatmap"
	"github.com/zeebo/errs"
	"golang.org/x/image/font/inconsolata"
)

// Options are all the ways you can configure the graph.
type Options struct {
	//
	// Data collection.
	//

	// Duration is the amount of history the graph represents from now.
	Duration time.Duration

	// Samples is many samples to take for a merged distribution.
	Samples int

	// Params are the parameters to the tdigest.
	Params tdigest.Params

	//
	// Heatmap
	//

	// Colors used for the heatmap.
	Colors []draw.Color

	// Width and Height for the heatmap. The axes are extra.
	Width, Height int
}

// Graph is a struct that draws graphs.
type Graph struct {
	opts Options

	now         int64
	stop_before int64
	earliest    dists.Dist

	merger *merger
}

// New constructs a Graph with the given options.
func New(opts Options) *Graph {
	now := time.Now().UnixNano()

	return &Graph{
		opts: opts,

		now:         now,
		stop_before: now - opts.Duration.Nanoseconds(),

		merger: newMerger(mergerOptions{
			Width:    opts.Width,
			Samples:  opts.Samples,
			Now:      now,
			Duration: opts.Duration,
			MergeOptions: merge.MergeOptions{
				Params: opts.Params,
			},
		}),
	}
}

// Now returns the now time the graph is using.
func (g *Graph) Now() int64 { return g.now }

// Push is a callback that should be sent to a Query function on a disk.Disk.
func (g *Graph) Push(ctx context.Context, start, end int64, buf []byte) (
	bool, error) {

	var rec data.Record
	if err := rec.Unmarshal(buf); err != nil {
		return false, errs.Wrap(err)
	}

	if g.earliest == nil {
		dist, err := dists.Load(rec)
		if err != nil {
			return false, errs.Wrap(err)
		}
		g.earliest = dist
	}

	if err := g.merger.Push(ctx, rec); err != nil {
		return false, errs.Wrap(err)
	}

	return end < g.stop_before, nil
}

// Finish returns the drawn graph after all of the Push calls have been made.
func (g *Graph) Finish(ctx context.Context) (*draw.RGB, error) {
	// draw the heatmap if we got some data
	canvas := draw.NewRGB(g.opts.Width, g.opts.Height)

	if g.earliest != nil {
		cols, err := g.merger.Finish(ctx)
		if err != nil {
			return nil, errs.Wrap(err)
		}

		hm := heatmap.New(heatmap.Options{
			Canvas: canvas,
			Colors: g.opts.Colors,
			Map:    g.earliest.CDF,
		})

		for _, col := range cols {
			hm.Draw(ctx, col)
		}
	}

	// create the axes:
	var labels []axis.Label

	// 1. the percentile axis on the left
	labels = labels[:0]
	for y := 0; y <= g.opts.Height-30; y += 30 {
		pos := float64(y) / float64(g.opts.Height)
		labels = append(labels, axis.Label{
			Position: pos,
			Text:     fmt.Sprintf("%0.2f", pos),
		})
	}
	labels = append(labels, axis.Label{
		Position: 1,
		Text:     "1.00",
	})

	left := axis.Draw(axis.Options{
		Face:     inconsolata.Regular8x16,
		Labels:   labels,
		Vertical: true,
		Length:   g.opts.Height + 1,
		Flip:     true,
	})

	// 2. the value axis on the right (if we have data)
	var right *draw.RGB

	if g.earliest != nil {
		labels = labels[:0]
		for y := 0; y <= g.opts.Height-30; y += 30 {
			pos := float64(y) / float64(g.opts.Height)
			labels = append(labels, axis.Label{
				Position: pos,
				Text:     fmt.Sprintf("%0.4f", g.earliest.Query(pos)),
			})
		}
		labels = append(labels, axis.Label{
			Position: 1,
			Text:     fmt.Sprintf("%0.4f", g.earliest.Query(1)),
		})
		right = axis.Draw(axis.Options{
			Face:     inconsolata.Regular8x16,
			Labels:   labels,
			Vertical: true,
			Length:   g.opts.Height + 1,
		})
	}

	// 3. draw the time axis on the bottom
	labels = labels[:0]

	// 3a. determine the "natural" duration unit for 100 px. this is going to
	// be the largest "natural" unit smaller than the chunk.
	chunk := g.opts.Duration / time.Duration(g.opts.Width/100)
	var natural time.Duration
	for _, unit := range naturalUnits {
		if unit < chunk {
			natural = unit
		} else {
			break
		}
	}

	// 3b. create the labels  from the truncated now
	x := int64(time.Duration(g.now).Truncate(natural))
	for x > g.stop_before {
		t := time.Unix(0, x)
		labels = append(labels, axis.Label{
			Position: 1 - float64(g.now-x)/float64(g.opts.Duration),
			Text:     t.Format("1/2 @ 15:04"),
		})
		x -= natural.Nanoseconds()
	}

	// 3c. reverse the labels so that they are in increasing order
	for i := 0; i < len(labels)/2; i++ {
		si := len(labels) - 1 - i
		labels[i], labels[si] = labels[si], labels[i]
	}

	bottom := axis.Draw(axis.Options{
		Face:     inconsolata.Regular8x16,
		Labels:   labels,
		Vertical: false,
		Length:   g.opts.Width,
	})

	// 4. combine all the rendered stuff
	width := left.Width + bottom.Width
	width_canvas := left.Width + canvas.Width
	if right != nil {
		width_canvas += right.Width
	}
	if width_canvas > width {
		width = width_canvas
	}
	height := canvas.Height + bottom.Height

	out := draw.NewRGB(width, height)

	// TODO(jeff): this is insanity!

	// 4a. copy the left axis
	for x := 0; x < left.Width; x++ {
		for y := 0; y < left.Height; y++ {
			copy(out.Raw(x, y), left.Raw(x, y))
		}
	}

	// 4b. copy the canvas
	for x := 0; x < canvas.Width; x++ {
		for y := 0; y < canvas.Height; y++ {
			copy(out.Raw(x+left.Width, y), canvas.Raw(x, y))
		}
	}

	// 4c. copy the bottom axis
	for x := 0; x < bottom.Width; x++ {
		for y := 0; y < bottom.Height; y++ {
			copy(out.Raw(x+left.Width, y+canvas.Height), bottom.Raw(x, y))
		}
	}

	// 4d. copy the right if it exists
	if right != nil {
		offset := left.Width + canvas.Width
		for x := 0; x < right.Width; x++ {
			for y := 0; y < right.Height; y++ {
				copy(out.Raw(x+offset, y), right.Raw(x, y))
			}
		}
	}

	return out, nil
}
