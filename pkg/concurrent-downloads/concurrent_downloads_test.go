package concurrentdownloads_test

import (
	"strings"
	"testing"

	. "github.com/shakefu/go-concurrency-experiments/pkg/concurrent-downloads"
)

func TestConcurrentDownloads(t *testing.T) {
	t.Run("it works", func(t *testing.T) {
		urls := []string{
			"https://example.com",
		}

		ctx := t.Context()
		data, err := DownloadAll(ctx, urls)

		if err != nil {
			t.Error(err)
		}

		if len(data) != len(urls) {
			t.Errorf("data is the wrong length: %v != %v", len(data), len(urls))
		}

		body := data["https://example.com"]
		if !strings.Contains(body, "<title>Example Domain</title>") {
			t.Log(body)
			t.Errorf("data content did not match expected page content")
		}
	})

	t.Run("it downloads multiple urls", func(t *testing.T) {
		urls := []string{
			"https://example.com",
			"https://httpstatuses.maor.io/200",
		}
		ctx := t.Context()
		data, err := DownloadAll(ctx, urls)

		if err != nil {
			t.Error(err)
		}

		if len(data) != len(urls) {
			t.Errorf("data is the wrong length: %v != %v", len(data), len(urls))
			return
		}

		body := data[urls[1]]
		if !strings.Contains(body, "OK") {
			t.Log(body)
			t.Errorf("data content did not match expected page content")
		}
	})

	t.Run("it fails if a url fails", func(t *testing.T) {
		urls := []string{
			"https://httpstatuses.maor.io/500",
			"https://httpstatuses.maor.io/200",
			"https://httpstatuses.maor.io/204",
			"https://httpstatuses.maor.io/400",
			"https://httpstatuses.maor.io/401",
			"https://httpstatuses.maor.io/403",
			"https://httpstatuses.maor.io/404",
		}
		ctx := t.Context()
		data, err := DownloadAll(ctx, urls)

		if err == nil {
			t.Error("expected error, got none")
		}

		if !strings.HasPrefix(err.Error(), "error fetching URL") {
			t.Error("expected error to start with 'error fetching URL'")
		}

		// Only one URL can possibly succeed
		if len(data) > 2 {
			t.Errorf("data is the wrong length: %v > %v", len(data), 2)
		}
	})
}
