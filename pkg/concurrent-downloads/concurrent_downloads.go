// Package concurrentdownloads provides a DownloadAll function for concurrently
// downloading URLs.
package concurrentdownloads

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// DownloadAll returns a map of {url:data}
func DownloadAll(ctx context.Context, urls []string) (map[string]string, error) {
	ctx, cancel := context.WithCancelCause(ctx)

	data := make(map[string]string, len(urls))
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	fetch := func(url string) {
		body, err := FetchURL(ctx, url)
		if err != nil {
			// TODO(shakefu): Decide on error behavior, for now, we cancel
			// everything, and fail out
			if err != context.Canceled && err != context.DeadlineExceeded {
				// Don't try to re-cancel if we're canceled already
				cancel(err)
			}
			return
		}
		mu.Lock()
		data[url] = string(body)
		mu.Unlock()
	}

	for _, url := range urls {
		wg.Go(func() { fetch(url) })
	}
	wg.Wait()

	return data, context.Cause(ctx)
}

func FetchURL(ctx context.Context, url string) ([]byte, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Make errors happen for testing, mostly
	if res.StatusCode > 399 {
		return nil, fmt.Errorf("error fetching URL: %s, status code: %d", url, res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
