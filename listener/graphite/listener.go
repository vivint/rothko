// Copyright (C) 2018. See AUTHORS.

package graphite

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"strconv"
	"sync"

	"github.com/vivint/rothko/data"
	"github.com/vivint/rothko/external"
	"github.com/zeebo/errs"
)

// Listener implements the listener.Listener for the graphite wire protocol.
type Listener struct {
	address string
}

// New returns a Listener that when Run will listen on the provided address.
func New(address string) *Listener {
	return &Listener{
		address: address,
	}
}

// Run listens on the address and writes all of the metrics to the writer.
func (l *Listener) Run(ctx context.Context, w *data.Writer) (err error) {
	lis, err := net.Listen("tcp", l.address)
	if err != nil {
		return errs.Wrap(err)
	}
	defer lis.Close()

	var wg sync.WaitGroup
	var errs = make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- handleListener(ctx, w, lis)
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		lis.Close()
		wg.Wait()
		return nil
	}
}

// handleListener accepts connections from the listener and spawns handlers
// for them.
func handleListener(ctx context.Context, w *data.Writer, lis net.Listener) (
	err error) {

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := handleConn(ctx, w, conn)
			if err != nil {
				external.Errorw("graphite connection error",
					"err", err.Error(),
				)
			}
		}()
	}
}

// handleConn handles lines from the connection and adds them to the writer.
func handleConn(ctx context.Context, w *data.Writer, conn net.Conn) (
	err error) {

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		conn.Close()
	}()

	// TODO(jeff): don't return an error when the conn is closed due to the
	// context.

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		err := handleLine(ctx, w, scanner.Bytes())
		if err != nil {
			external.Errorw("invalid graphite line",
				"line", scanner.Text(),
				"peer", conn.RemoteAddr().String(),
				"err", err.Error(),
			)
		}
	}
	return scanner.Err()
}

// handleLine adds the graphite data in the line to the writer.
func handleLine(ctx context.Context, w *data.Writer, line []byte) (err error) {
	fields := bytes.Split(line, []byte{' '})
	if len(fields) != 3 {
		return errs.New("bad number of fields: %d", len(fields))
	}

	value, err := strconv.ParseFloat(string(fields[1]), 64)
	if err != nil {
		return errs.Wrap(err)
	}

	w.Add(ctx, string(fields[0]), value, nil)
	return nil
}
