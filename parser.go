package main

import (
	"database/sql"
	"log"
	"strings"
	"time"
)

func parsePubDate(s string) sql.NullTime {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	layouts := []string{
		time.RFC1123Z,                     // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,                      // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC822Z,                      // "02 Jan 06 15:04 -0700"
		time.RFC822,                       // "02 Jan 06 15:04 MST"
		time.RFC3339,                      // ISO 8601
		"Mon, 02 Jan 2006 15:04:05 -0700", // common variants
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}
	log.Printf("unparsed PubDate: %q", s)
	return sql.NullTime{Valid: false}
}
