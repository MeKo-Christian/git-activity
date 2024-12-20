package internal

import (
	"fmt"
	"log/slog"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func GenerateGroupedCharts(combinedActivity *CombinedCommitActivity, mode string, outputPrefix, format string) error {
	// Extract repository names and activity data
	repoNames := []string{}
	weekdays := make(map[string][]int)
	hours := make(map[string][]int)
	months := make(map[string][]int)
	weeks := make(map[string][]int)

	for _, repoActivity := range combinedActivity.Repos {
		repoNames = append(repoNames, repoActivity.RepoName)
		weekdays[repoActivity.RepoName] = repoActivity.Activity.Weekdays[:]
		hours[repoActivity.RepoName] = repoActivity.Activity.Hours[:]
		months[repoActivity.RepoName] = repoActivity.Activity.Months[:]
		weeks[repoActivity.RepoName] = repoActivity.Activity.Weeks[:]
	}

	// Generate grouped bar charts
	err := CreateGroupedChart("Combined Commits by Weekday", weekdays, repoNames, WeekdayLabels(), fmt.Sprintf("%s_by_weekday.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating grouped weekday chart: %w", err)
	}

	err = CreateGroupedChart("Combined Commits by Hour", hours, repoNames, HourLabels(), fmt.Sprintf("%s_by_hour.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating grouped hourly chart: %w", err)
	}

	err = CreateGroupedChart("Combined Commits by Month", months, repoNames, MonthLabels(), fmt.Sprintf("%s_by_month.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating grouped monthly chart: %w", err)
	}

	err = CreateGroupedChart("Combined Commits by Week", weeks, repoNames, WeekLabels(), fmt.Sprintf("%s_by_week.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating grouped weekly chart: %w", err)
	}

	return nil
}

func CreateGroupedChart(title string, data map[string][]int, repoNames []string, labels []string, filename string) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Categories"
	p.Y.Label.Text = "Commits"

	barWidth := vg.Points(20)

	for i, repoName := range repoNames {
		values := plotter.Values{}
		for _, val := range data[repoName] {
			values = append(values, float64(val))
		}

		bar, err := plotter.NewBarChart(values, barWidth)
		if err != nil {
			return fmt.Errorf("error creating bar chart for %s: %w", repoName, err)
		}

		bar.Offset = vg.Length(i) * barWidth
		p.Add(bar)
	}

	p.NominalX(labels...)
	return p.Save(15*vg.Inch, 6*vg.Inch, filename)
}

func GenerateCharts(combinedActivity *CombinedCommitActivity, mode string, outputPrefix, format string) error {
	chartTitle := "Commits"
	if mode == "lines" {
		chartTitle = "Lines Changed"
	}

	overallActivity := &CommitActivity{}
	for _, repoActivity := range combinedActivity.Repos {
		overallActivity.Combine(repoActivity.Activity)
	}

	slog.Info("Generating standard charts", "output_prefix", outputPrefix, "format", format)

	// Generate standard charts
	err := CreateBarChart(overallActivity.Weekdays[:], WeekdayLabels(), fmt.Sprintf("Combined %s by Weekday", chartTitle), fmt.Sprintf("%s_by_weekday.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating weekday chart: %w", err)
	}

	err = CreateBarChart(overallActivity.Hours[:], HourLabels(), fmt.Sprintf("Combined %s by Hour", chartTitle), fmt.Sprintf("%s_by_hour.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating hourly chart: %w", err)
	}

	err = CreateBarChart(overallActivity.Months[:], MonthLabels(), fmt.Sprintf("Combined %s by Month", chartTitle), fmt.Sprintf("%s_by_month.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating monthly chart: %w", err)
	}

	err = CreateBarChart(overallActivity.Weeks[:], WeekLabels(), fmt.Sprintf("Combined %s by Week", chartTitle), fmt.Sprintf("%s_by_week.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating weekly chart: %w", err)
	}

	return nil
}
