// Copyright (C) 2018. See AUTHORS.

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/load"
	"github.com/spacemonkeygo/rothko/database/files"
	"github.com/spacemonkeygo/rothko/internal/junk"
	"github.com/zeebo/errs"
)

func runLatest(ctx context.Context, di *files.DB, metric string) error {
	start, end, data, err := di.QueryLatest(ctx, metric, nil)
	if err != nil {
		return err
	}
	return printData(start, end, data)
}

func printData(start, end int64, buf []byte) error {
	var rec data.Record
	if err := rec.Unmarshal(buf); err != nil {
		return errs.Wrap(err)
	}
	dist, err := load.Load(rec)
	if err != nil {
		return errs.Wrap(err)
	}

	tw := junk.NewTabbed(os.Stdout)
	tw.Write("start:", time.Unix(0, start).Format(time.RFC1123), fmt.Sprintf("(%d)", start))
	tw.Write("end:", time.Unix(0, end).Format(time.RFC1123), fmt.Sprintf("(%d)", end))
	tw.Write("obs:", fmt.Sprint(rec.Observations))
	tw.Write("kind:", rec.Kind)
	tw.Write("data:", fmt.Sprintf("%x", rec.Distribution))
	tw.Write("min:", fmt.Sprint(rec.Min), fmt.Sprintf("%x", rec.MinId))
	tw.Write("max:", fmt.Sprint(rec.Max), fmt.Sprintf("%x", rec.MaxId))
	tw.Write("merged:", fmt.Sprint(rec.Merged))

	for x := 0.0; x <= 1.0; x += 1.0 / 32 {
		val := dist.Query(x)
		tw.Write(fmt.Sprintf("%0.2f:", x), fmt.Sprintf("%0.6f", val))
	}
	tw.Flush()

	return errs.Wrap(tw.Error())
}
