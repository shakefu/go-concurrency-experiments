/*
* Package log_parse provides functions for parsing log files.
*
Web server log sample:
[02/Nov/2018:21:46:31 +0000] PUT /users/12345/locations HTTP/1.1 204 iphone-3
[02/Nov/2018:21:46:31 +0000] PUT /users/6098/locations HTTP/1.1 204 iphone-3
[02/Nov/2018:21:46:32 +0000] PUT /users/3911/locations HTTP/1.1 204 moto-x
[02/Nov/2018:21:46:33 +0000] PUT /users/9933/locations HTTP/1.1 404 moto-x
[02/Nov/2018:21:46:33 +0000] PUT /users/3911/locations HTTP/1.1 500 moto-x
[02/Nov/2018:21:46:34 +0000] GET /rides/9943222/status HTTP/1.1 200 moto-x
[02/Nov/2018:21:46:34 +0000] POST /rides HTTP/1.1 202 iphone-2
[02/Nov/2018:21:46:35 +0000] POST /users HTTP/1.1 202 iphone-5
[02/Nov/2018:21:46:35 +0000] POST /rides HTTP/1.1 202 iphone-5
[02/Nov/2018:21:46:37 +0000] POST /rides HTTP/1.1 202 iphone-4
[02/Nov/2018:21:46:38 +0000] GET /users/994/ride/16 HTTP/1.1 200 iphone-5
[02/Nov/2018:21:46:39 +0000] POST /users HTTP/1.1 202 iphone-3
[02/Nov/2018:21:46:40 +0000] PUT /users/8384721/locations HTTP/1.1 204 iphone-3
[02/Nov/2018:21:46:41 +0000] GET /users/342111 HTTP/1.1 200 iphone-5
[02/Nov/2018:21:46:42 +0000] GET /users/9933 HTTP/1.1 200 iphone-5
[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 200 iphone-5

Take a log and output a table representing the number of occurrences of events in the
log.

Where the event is defined as (Method + Endpoint + HttpStatusCode).

Order by Count, descending:

Method |             Endpoint | Code || Count
=============================================

	PUT   |   /users/#/locations | 204  ||  4
	POST  |               /rides | 202  ||  3
	GET   |             /users/# | 200  ||  2
	POST  |               /users | 202  ||  2
	PUT   |   /users/#/locations | 500  ||  1
	GET   |      /prices/#/geo/# | 200  ||  1
	PUT   |   /users/#/locations | 404  ||  1
	GET   |      /rides/#/status | 200  ||  1
	GET   |      /users/#/ride/# | 200  ||  1
*/
package log_parse

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss/v2/table"
)

var ErrInvalidLogFormat = errors.New("invalid log format")

type Event struct {
	Method     string
	Endpoint   string
	HttpStatus int
}

type EventCount struct {
	Event
	Count int
}

// Parse a log line and return structured data.
//
// Expects logs like:
//
//	[02/Nov/2018:21:46:43 +0000] GET /prices/20180103/geo/12 HTTP/1.1 200 iphone-5
//	[DD/mmm/YYYY:HH:MM:SS +0000] METHOD URI HTTP/1.1 STATUS USERAGENT
func ParseLine(record string) (Event, error) {
	pattern := `\[[^[]+\] ([A-Z]+) (.+) HTTP/1.1 (\d+) (.+)`
	// Build a regex
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(record)
	if match != nil {
		// timestamp := match[0]
		method := match[1]
		endpoint := match[2]
		statusStr := match[3]
		// userAgent := match[4]

		statusCode, _ := strconv.Atoi(statusStr)

		return Event{
			Method:     method,
			Endpoint:   endpoint,
			HttpStatus: statusCode,
		}, nil
	}
	return Event{}, ErrInvalidLogFormat
}

// CountEvents counts the number of events in a list of logs.
func CountEvents(logs []string) []*EventCount {
	events := make(map[Event]*EventCount)
	for _, log := range logs {
		event, err := ParseLine(log)
		if err != nil {
			continue
		}
		if _, ok := events[event]; !ok {
			events[event] = &EventCount{
				Event: event,
				Count: 0,
			}
		}
		ev := events[event]
		ev.Count++
		events[event] = ev
	}

	// Get all the eventcounts and sort them
	counts := slices.Collect(maps.Values(events))
	sort.Slice(counts, func(i, j int) bool {
		// Implementing deterministic sorting, not really necessary but improves
		// testability
		if counts[i].Count == counts[j].Count {
			if counts[i].Event.Method == counts[j].Event.Method {
				return counts[i].Event.Endpoint < counts[j].Event.Endpoint
			}
			return counts[i].Event.Method < counts[j].Event.Method
		}
		return counts[i].Count > counts[j].Count
	})

	return counts
}

// ParseLog parses a text containing log lines and returns the count of events.
func ParseLog(text string) []*EventCount {
	lines := strings.Split(text, "\n")
	return CountEvents(lines)
}

// Output the counts of events in a nicely formatted table using `lipgloss`
func FormatCounts(counts []*EventCount) string {
	table := table.New().
		Headers("Method", "Endpoint", "Count").
		Rows(func() [][]string {
			rows := make([][]string, len(counts))
			for i, count := range counts {
				rows[i] = []string{
					count.Event.Method,
					count.Event.Endpoint,
					fmt.Sprint(count.Count),
				}
			}
			return rows
		}()...)
	return table.Render()
}
