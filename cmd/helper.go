package cmd

import (
	"log"
	"time"
)

func parseDateRange(startStr, endStr string) (time.Time, time.Time) {
	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			log.Fatalf("Invalid start date format: %v", err)
		}
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			log.Fatalf("Invalid end date format: %v", err)
		}
	}

	if !start.IsZero() && !end.IsZero() && start.After(end) {
		log.Fatalf("Start date cannot be after end date.")
	}

	return start, end
}
