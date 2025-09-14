package log_parse_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	. "github.com/shakefu/go-concurrency-experiments/pkg/log-parse"
)

func TestParseLine(t *testing.T) {
	// Test that the function returns an error for invalid log format
	invalidLog := "[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 200 iphone-5"
	event, err := ParseLine(invalidLog)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		t.Fail()
	}
	expectedEvent := Event{
		Method:     "GET",
		Endpoint:   "/prices/20180103/geo/12",
		HttpStatus: 200,
	}
	if event != expectedEvent {
		t.Errorf("Expected %v, got %v", expectedEvent, event)
	}

	// Test that the function returns the correct Event struct for a valid log
	validLog := "[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 200 iphone-5"
	event, err = ParseLine(validLog)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedEvent = Event{
		Method:     "GET",
		Endpoint:   "/prices/20180103/geo/12",
		HttpStatus: 200,
	}
	if event != expectedEvent {
		t.Errorf("Expected %v, got %v", expectedEvent, event)
	}

	// Test that the function returns the correct Event struct for a valid log with different status code
	validLog2 := "[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 404 iphone-5"
	event2, err := ParseLine(validLog2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedEvent2 := Event{
		Method:     "GET",
		Endpoint:   "/prices/20180103/geo/12",
		HttpStatus: 404,
	}
	if event2 != expectedEvent2 {
		t.Errorf("Expected %v, got %v", expectedEvent2, event2)
	}

	entries := []string{
		"[02/Nov/2018:21:46:31 +0000] PUT /users/12345/locations HTTP/1.1 204 iphone-3",
		"[02/Nov/2018:21:46:31 +0000] PUT /users/6098/locations HTTP/1.1 204 iphone-3",
		"[02/Nov/2018:21:46:32 +0000] PUT /users/3911/locations HTTP/1.1 204 moto-x",
		"[02/Nov/2018:21:46:33 +0000] PUT /users/9933/locations HTTP/1.1 404 moto-x",
		"[02/Nov/2018:21:46:33 +0000] PUT /users/3911/locations HTTP/1.1 500 moto-x",
		"[02/Nov/2018:21:46:34 +0000] GET /rides/9943222/status HTTP/1.1 200 moto-x",
		"[02/Nov/2018:21:46:34 +0000] POST /rides HTTP/1.1 202 iphone-2",
		"[02/Nov/2018:21:46:35 +0000] POST /users HTTP/1.1 202 iphone-5",
		"[02/Nov/2018:21:46:35 +0000] POST /rides HTTP/1.1 202 iphone-5",
		"[02/Nov/2018:21:46:37 +0000] POST /rides HTTP/1.1 202 iphone-4",
		"[02/Nov/2018:21:46:38 +0000] GET /users/994/ride/16 HTTP/1.1 200 iphone-5",
		"[02/Nov/2018:21:46:39 +0000] POST /users HTTP/1.1 202 iphone-3",
		"[02/Nov/2018:21:46:40 +0000] PUT /users/8384721/locations HTTP/1.1 204 iphone-3",
		"[02/Nov/2018:21:46:41 +0000] GET /users/342111 HTTP/1.1 200 iphone-5",
		"[02/Nov/2018:21:46:42 +0000] GET /users/9933 HTTP/1.1 200 iphone-5",
		"[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 200 iphone-5",
	}

	expected := [][2]string{
		[2]string{"PUT", "/users/12345/locations"},
		[2]string{"PUT", "/users/6098/locations"},
		[2]string{"PUT", "/users/3911/locations"},
		[2]string{"PUT", "/users/9933/locations"},
		[2]string{"PUT", "/users/3911/locations"},
		[2]string{"GET", "/rides/9943222/status"},
		[2]string{"POST", "/rides"},
		[2]string{"POST", "/users"},
		[2]string{"POST", "/rides"},
		[2]string{"POST", "/rides"},
		[2]string{"GET", "/users/994/ride/16"},
		[2]string{"POST", "/users"},
		[2]string{"PUT", "/users/8384721/locations"},
		[2]string{"GET", "/users/342111"},
		[2]string{"GET", "/users/9933"},
		[2]string{"GET", "/prices/20180103/geo/12"},
	}

	for i := range entries {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			entry := entries[i]
			expected := expected[i]

			event, err := ParseLine(entry)
			if err != nil {
				t.Errorf("Error parsing line: %v", err)
			}
			if event.Method != expected[0] || event.Endpoint != expected[1] {
				t.Errorf("Expected %s %s, got %s %s", expected[0], expected[1], event.Method, event.Endpoint)
			}
		})
	}
}

func TestCountEvents(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []*EventCount
	}{
		{
			name: "single event",
			input: []string{
				"[02/Nov/2018:21:46:42 +0000] GET /users/9933 HTTP/1.1 200 iphone-5",
			},
			expected: []*EventCount{
				{Event{Method: "GET", Endpoint: "/users/9933", HttpStatus: 200}, 1},
			},
		},
		{
			name: "multiple events",
			input: []string{
				"[02/Nov/2018:21:46:42 +0000] GET /users/9933 HTTP/1.1 200 iphone-5",
				"[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 200 iphone-5",
			},
			expected: []*EventCount{
				{Event{Method: "GET", Endpoint: "/prices/20180103/geo/12", HttpStatus: 200}, 1},
				{Event{Method: "GET", Endpoint: "/users/9933", HttpStatus: 200}, 1},
			},
		},
		{
			name: "multiple events with different status codes",
			input: []string{
				"[02/Nov/2018:21:46:42 +0000] GET /users/9933 HTTP/1.1 200 iphone-5",
				"[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 404 iphone-5",
			},
			expected: []*EventCount{
				{Event{Method: "GET", Endpoint: "/prices/20180103/geo/12", HttpStatus: 404}, 1},
				{Event{Method: "GET", Endpoint: "/users/9933", HttpStatus: 200}, 1},
			},
		},
		{
			name: "multiple events with same endpoint but different methods",
			input: []string{
				"[02/Nov/2018:21:46:42 +0000] GET /users/9933 HTTP/1.1 200 iphone-5",
				"[02/Nov/2018:21:46:43 +0000] POST /users/9933 HTTP/1.1 200 iphone-5",
			},
			expected: []*EventCount{
				{Event{Method: "GET", Endpoint: "/users/9933", HttpStatus: 200}, 1},
				{Event{Method: "POST", Endpoint: "/users/9933", HttpStatus: 200}, 1},
			},
		},
		{
			name: "multiple identical calls",
			input: []string{
				"[02/Nov/2018:21:46:42 +0000] GET /users/9933 HTTP/1.1 200 iphone-5",
				"[02/Nov/2018:21:46:43 +0000] GET /users/9933 HTTP/1.1 200 iphone-5",
			},
			expected: []*EventCount{
				{Event{Method: "GET", Endpoint: "/users/9933", HttpStatus: 200}, 2},
			},
		},
		{
			name: "multiple same calls with varying user agent",
			input: []string{
				"[02/Nov/2018:21:46:42 +0000] GET /users/9933 HTTP/1.1 200 iphone-5",
				"[02/Nov/2018:21:46:43 +0000] GET /users/9933 HTTP/1.1 200 android-7",
			},
			expected: []*EventCount{
				{Event{Method: "GET", Endpoint: "/users/9933", HttpStatus: 200}, 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CountEvents(tt.input)
			for i := range tt.expected {
				if !reflect.DeepEqual(actual[i], tt.expected[i]) {
					t.Errorf("Expected %v, got %v", tt.expected[i], actual[i])
				}
			}
		})
	}
}

func TestParseLineRegex(t *testing.T) {
	pattern := `\[([^[])+\] ([A-Z]+) (.+) HTTP/1.1 (\d+) (.+)`
	// Build a regex
	re := regexp.MustCompile(pattern)

	record := "[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 200 iphone-5"

	match := re.FindStringSubmatch(record)

	if len(match) != 6 {
		t.Errorf("Expected 6 matches, got %d", len(match))
	}
}

func TestFormatCounts(t *testing.T) {
	counts := []*EventCount{
		{Event{Method: "GET", Endpoint: "/users/9933", HttpStatus: 200}, 2},
	}

	expected := `╭──────┬───────────┬─────╮
│Method│Endpoint   │Count│
├──────┼───────────┼─────┤
│GET   │/users/9933│2    │
╰──────┴───────────┴─────╯`

	actual := FormatCounts(counts)

	if actual != expected {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
