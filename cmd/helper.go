package cmd

import (
	"bufio"
	"git-activity/internal"
	"log"
	"os"
	"strings"
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

// ParseDeveloperAliases parses a file into a map of aliases to developer names
func parseDeveloperAliases(filename string) (internal.DeveloperAliases, error) {
	aliases := internal.DeveloperAliases{}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue // Skip invalid lines
		}

		name := parts[0]
		for _, alias := range parts[1:] {
			aliases[strings.ToLower(strings.TrimSpace(alias))] = name
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return aliases, nil
}
