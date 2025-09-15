package orderedlogs_test

import (
	"bytes"
	"testing"
	"time"

	. "github.com/shakefu/go-concurrency-experiments/pkg/ordered-logs"
)

func TestOrderedLogs(t *testing.T) {
	waitForDone := func(done chan bool) {
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout")
		}
	}

	t.Run("it works", func(t *testing.T) {
		buf := &bytes.Buffer{}
		done := make(chan bool)
		o := &OrderedLogs{}
		o.Out = buf

		go func() {
			o.PrintFirst("one")
			o.PrintSecond("two")
			o.PrintThird("three")

			close(done)
		}()

		waitForDone(done)

		o.Wait()

		b := buf.String()
		if b != "one\ntwo\nthree\n" {
			t.Errorf("unexpected output: '%v'", b)
		}
	})

	t.Run("out of order", func(t *testing.T) {
		buf := &bytes.Buffer{}
		done := make(chan bool)
		o := &OrderedLogs{}
		o.Out = buf

		go func() {
			o.PrintThird("three")
			o.PrintSecond("two")
			o.PrintFirst("one")

			close(done)
		}()

		waitForDone(done)

		o.Wait()

		b := buf.String()
		e := "one\ntwo\nthree\n"
		if b != e {
			t.Errorf("expected: '%v', got '%v'", e, b)
		}
	})

	t.Run("out of space", func(t *testing.T) {
		buf := &bytes.Buffer{}
		done := make(chan bool)
		o := &OrderedLogs{}
		o.Out = buf
		o.Size = 1

		go func() {
			defer func() {
				close(done)
				r := recover()
				if r == nil {
					t.Errorf("The code did not panic")
				}
				if r != "Out of space" {
					t.Errorf("expected 'Out of space', got '%v'", r)
				}
			}()

			o.PrintThird("three")
			o.PrintSecond("two")
			o.PrintFirst("one")

			o.PrintSecond("2")
			o.PrintFirst("1")
			o.PrintThird("3")
		}()

		waitForDone(done)

		// This should be idempotent
		o.Wait()
	})

	t.Run("multiple statements", func(t *testing.T) {
		buf := &bytes.Buffer{}
		done := make(chan bool)
		o := &OrderedLogs{}
		o.Out = buf

		go func() {
			o.PrintThird("three")
			o.PrintSecond("two")
			o.PrintFirst("one")

			o.PrintSecond("2")
			o.PrintFirst("1")
			o.PrintThird("3")

			o.PrintThird("third")
			o.PrintThird("three")
			o.PrintFirst("first")
			o.PrintSecond("second")
			o.PrintFirst("one")
			o.PrintSecond("two")

			close(done)
		}()

		waitForDone(done)

		o.Wait()

		b := buf.String()
		e := "one\ntwo\nthree\n1\n2\n3\nfirst\nsecond\nthird\none\ntwo\nthree\n"
		if b != e {
			t.Errorf("expected: '%v', got '%v'", e, b)
		}
	})
}
