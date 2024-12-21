package internal

import (
	"fmt"
	"log/slog"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Utility methods for labels
func WeekdayLabels() []string {
	return []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
}

func HourLabels() []string {
	labels := make([]string, 24)
	for i := 0; i < 24; i++ {
		labels[i] = fmt.Sprintf("%02d:00", i)
	}
	return labels
}

func MonthLabels() []string {
	return []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
}

func WeekLabels() []string {
	labels := make([]string, 53)
	for i := 0; i < 53; i++ {
		labels[i] = fmt.Sprintf("Week %d", i)
	}
	return labels
}

func GenerateCharts(combinedActivity *CombinedCommitActivity, mode string, outputPrefix, format string, aliases DeveloperAliases) error {
	chartTitle := "Commits"
	if mode == "lines" {
		chartTitle = "Lines Changed"
	}

	overallActivity := NewCommitActivity()
	for _, repoActivity := range combinedActivity.Repos {
		overallActivity.Combine(repoActivity.Activity)
	}

	// Aggregate data for non-grouped charts
	weekdays := aggregateMapData(overallActivity.Weekdays, 7)
	hours := aggregateMapData(overallActivity.Hours, 24)
	months := aggregateMapData(overallActivity.Months, 12)
	weeks := aggregateMapData(overallActivity.Weeks, 53)

	// Generate standard charts
	err := CreateBarChart(weekdays, WeekdayLabels(), fmt.Sprintf("Combined %s by Weekday", chartTitle), fmt.Sprintf("%s_by_weekday.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating weekday chart: %w", err)
	}

	err = CreateBarChart(hours, HourLabels(), fmt.Sprintf("Combined %s by Hour", chartTitle), fmt.Sprintf("%s_by_hour.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating hourly chart: %w", err)
	}

	err = CreateBarChart(months, MonthLabels(), fmt.Sprintf("Combined %s by Month", chartTitle), fmt.Sprintf("%s_by_month.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating monthly chart: %w", err)
	}

	err = CreateBarChart(weeks, WeekLabels(), fmt.Sprintf("Combined %s by Week", chartTitle), fmt.Sprintf("%s_by_week.%s", outputPrefix, format))
	if err != nil {
		return fmt.Errorf("error creating weekly chart: %w", err)
	}

	return nil
}

func GenerateGroupedCharts(
	combinedActivity *CombinedCommitActivity,
	mode, stacking, outputPrefix, format string,
	aliases DeveloperAliases,
) error {
	slog.Info("Generating grouped charts", "output_prefix", outputPrefix, "format", format, "mode", mode, "stacking", stacking)

	// Labels for each category
	categories := []struct {
		activityKey func(activity *CommitActivity) map[string][]int
		labels      []string
		filename    string
		title       string
	}{
		{func(a *CommitActivity) map[string][]int { return a.Weekdays }, WeekdayLabels(), "by_weekday", "Activity by Weekday"},
		{func(a *CommitActivity) map[string][]int { return a.Hours }, HourLabels(), "by_hour", "Activity by Hour"},
		{func(a *CommitActivity) map[string][]int { return a.Months }, MonthLabels(), "by_month", "Activity by Month"},
		{func(a *CommitActivity) map[string][]int { return a.Weeks }, WeekLabels(), "by_week", "Activity by Week"},
	}

	for _, category := range categories {
		// Aggregate grouped data based on stacking
		groupedData := make(map[string]map[string]int)
		for _, repoActivity := range combinedActivity.Repos {
			data := category.activityKey(repoActivity.Activity)
			groupBy := "repo"
			if stacking == "byDev" {
				groupBy = "developer"
			}
			categoryGrouped := prepareGroupedData(data, groupBy, repoActivity.RepoName, category.labels)
			for key, value := range categoryGrouped {
				if groupedData[key] == nil {
					groupedData[key] = make(map[string]int)
				}
				for subKey, subValue := range value {
					groupedData[key][subKey] += subValue
				}
			}
		}

		// Generate chart for the current category
		err := CreateStackedBarChart(
			groupedData,
			category.labels,
			fmt.Sprintf("%s (%s)", category.title, stacking),
			fmt.Sprintf("%s_%s.%s", outputPrefix, category.filename, format),
		)
		if err != nil {
			return fmt.Errorf("error creating stacked chart for %s: %w", category.title, err)
		}
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

// CreateBarChart creates a bar chart from the given data
func CreateBarChart(values []int, labels []string, title, filename string) error {
	pts := make(plotter.Values, len(values))
	for i, v := range values {
		pts[i] = float64(v)
	}

	p := plot.New()
	p.Title.Text = title
	bar, err := plotter.NewBarChart(pts, vg.Points(20))
	if err != nil {
		return fmt.Errorf("could not create bar chart: %w", err)
	}

	bar.LineStyle.Width = vg.Length(0)
	bar.Color = colorPalette[1%len(colorPalette)]

	p.Add(bar)
	p.NominalX(labels...)

	// Save the chart to the specified file format (determined by file extension)
	err = p.Save(10*vg.Inch, 4*vg.Inch, filename)
	if err != nil {
		return fmt.Errorf("could not save chart: %w", err)
	}

	return nil
}

// CreateStackedBarChart creates a stacked bar chart from the given data
func CreateStackedBarChart(
	data map[string]map[string]int,
	labels []string,
	title, filename string,
) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Categories"
	p.Y.Label.Text = "Values"

	barWidth := vg.Points(20)
	categoryKeys := sortedKeys(data)

	for j, category := range categoryKeys {
		stack := plotter.Values{}
		for _, label := range labels {
			stack = append(stack, float64(data[category][label]))
		}

		bars, err := plotter.NewBarChart(stack, barWidth)
		if err != nil {
			return fmt.Errorf("could not create bar chart for %s: %w", category, err)
		}

		bars.LineStyle.Width = vg.Length(0)
		bars.Color = colorPalette[j%len(colorPalette)]

		p.Add(bars)
	}

	p.NominalX(labels...)
	return p.Save(15*vg.Inch, 6*vg.Inch, filename)
}

func sortedKeys[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func aggregateMapData(data map[string][]int, size int) []int {
	aggregated := make([]int, size)
	for developer, array := range data {
		if len(array) == 0 {
			fmt.Printf("Warning: Developer %s has empty data.\n", developer)
			continue
		}
		for i, value := range array {
			if i < size {
				aggregated[i] += value
			}
		}
	}
	return aggregated
}

func prepareGroupedData(data map[string][]int, groupBy, identifier string, labels []string) map[string]map[string]int {
	groupedData := make(map[string]map[string]int)

	for developer, values := range data {
		key := identifier
		if groupBy == "developer" {
			key = developer
		}
		for i, value := range values {
			if i >= len(labels) {
				continue
			}
			category := labels[i]
			if groupedData[key] == nil {
				groupedData[key] = make(map[string]int)
			}
			groupedData[key][category] += value
		}
	}

	return groupedData
}
