package gql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDaysOld covers the unparseable dates that used to fall through to the
// zero time, reporting an account as roughly 739,000 days old and pinning the
// account age trust factor at its maximum.
func TestDaysOld(t *testing.T) {
	tests := map[string]struct {
		createdAt string

		expectedDays float64
	}{
		"empty date":       {createdAt: "", expectedDays: 0},
		"malformed date":   {createdAt: "not a date", expectedDays: 0},
		"wrong format":     {createdAt: "2019-01-02", expectedDays: 0},
		"one day old":      {createdAt: time.Now().UTC().AddDate(0, 0, -1).Format(iso8601Format), expectedDays: 1},
		"one year old":     {createdAt: time.Now().UTC().AddDate(-1, 0, 0).Format(iso8601Format), expectedDays: 365},
		"created just now": {createdAt: time.Now().UTC().Format(iso8601Format), expectedDays: 0},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			t.Parallel()

			days := User{CreatedAt: test.createdAt}.DaysOld()

			assert.InDelta(t, test.expectedDays, days, 1.5)
		})
	}
}

func TestMetaDataCursor(t *testing.T) {
	assert.Empty(t, metaData{}.cursor(), "an empty edge list has no cursor to paginate on")
	assert.Equal(t, "last", metaData{{Cursor: "first"}, {Cursor: "last"}}.cursor())
}

func TestContribFilePaginationIsUnique(t *testing.T) {
	seen := make(map[string]bool)

	for _, cursor := range []string{"", "abc", "abc-2019"} {
		for year := 2013; year <= 2026; year++ {
			key := contribFilePagination(cursor, year)

			assert.False(t, seen[key], "cache key %q is used for more than one page", key)
			seen[key] = true
		}
	}

	assert.Len(t, seen, 3*14, "expected one cache key per cursor and year")
}
