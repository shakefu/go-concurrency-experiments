// Package orderedlogs implements a demonstration concurrent logs model.
//
// Each of the methods [PrintFirst], [PrintSecond], [PrintThird] guarantees that
// one message from each source will be printed, in order, before the next group
// is printed.
package orderedlogs

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

// OrderedLogs implements an ordered log printer
type OrderedLogs struct {
	first  chan *string
	second chan *string
	third  chan *string
	wg     sync.WaitGroup
	Size   int
	Out    io.Writer
	init   atomic.Bool
	wait   atomic.Bool
}

// New initializes a new OrderedLogs
func (o *OrderedLogs) New() *OrderedLogs {
	if o == nil {
		o = &OrderedLogs{}
	}

	if o.init.Load() {
		// Never initialize twice
		// TODO(shakefu): Output some kind of warning
		return o
	}

	// Default to stdout
	if o.Out == nil {
		o.Out = os.Stdout
	}

	// Default to 512 which fits in one page of memory
	if o.Size == 0 {
		o.Size = 512
	}
	o.first = make(chan *string, o.Size)
	o.second = make(chan *string, o.Size)
	o.third = make(chan *string, o.Size)
	o.wg = sync.WaitGroup{}
	o.init.Store(true)

	// Launch the loop to print
	o.wg.Add(1)
	go o.printLoop()

	return o
}

// Wait will wait on the printLoop to finish its work
func (o *OrderedLogs) Wait() {
	if o == nil || !o.init.Load() {
		panic("cannot wait on nil/uninitialized OrderedLogs")
	}

	if o.wait.Load() {
		// We're already waiting, so we just hang out instead of closing things
		o.wg.Wait()
		return
	}

	o.wait.Store(true)
	o.wg.Wait()
	o.Close()
}

// Close will close all the channels, without waiting for them to print
func (o *OrderedLogs) Close() {
	if o == nil || !o.init.Load() {
		return
	}

	select {
	case <-o.first:
		close(o.first)
	default:
		break
	}

	select {
	case <-o.second:
		close(o.second)
	default:
		break
	}

	select {
	case <-o.third:
		close(o.third)
	default:
		break
	}
}

// printLoop will loop wait and print one of each log type, in order until it is
// blocked, waiting for a log.
func (o *OrderedLogs) printLoop() {
	if o == nil || !o.init.Load() {
		return
	}

	print := func(s *string) {
		fmt.Fprintln(o.Out, *s)
	}

	for {
		var ok bool
		// Get the statements in order, if we have any
		firstLog, ok := <-o.first
		if !ok {
			break
		}

		if o.wait.Load() && (len(o.second) == 0 || len(o.third) == 0) {
			break
		}

		secondLog, ok := <-o.second
		if !ok {
			break
		}
		if o.wait.Load() && len(o.third) == 0 {
			break
		}

		thirdLog, ok := <-o.third
		if !ok {
			break
		}

		// If we're here, we successfully have a whole message
		print(firstLog)
		print(secondLog)
		print(thirdLog)

		if o.wait.Load() && len(o.first) == 0 {
			break
		}
	}

	o.wg.Done()
}

// PrintFirst will queue the first log for printing.
func (o *OrderedLogs) PrintFirst(log string) {
	if o == nil {
		// TODO(shakefu): Print a warning
		return
	}

	// We allow self-init for convenience magic
	if !o.init.Load() {
		o.New()
	}

	select {
	case o.first <- &log:
		// Successfully queued the log for printing, we're good
		return
	default:
		// We ran out of space! Panic for now.
		panic("Out of space")
	}
}

// PrintSecond will queue the second log for printing.
func (o *OrderedLogs) PrintSecond(log string) {
	if o == nil {
		// TODO(shakefu): Print a warning
		return
	}

	// We allow self-init for convenience magic
	if !o.init.Load() {
		o.New()
	}

	select {
	case o.second <- &log:
		// Successfully queued the log for printing, we're good
		return
	default:
		// We ran out of space! Panic for now.
		panic("Out of space")
	}
}

// PrintThird will queue the third log for printing.
func (o *OrderedLogs) PrintThird(log string) {
	if o == nil {
		// TODO(shakefu): Print a warning
		return
	}

	// We allow self-init for convenience magic
	if !o.init.Load() {
		o.New()
	}

	select {
	case o.third <- &log:
		// Successfully queued the log for printing, we're good
		return
	default:
		// We ran out of space! Panic for now.
		panic("Out of space")
	}
}
