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

func GenerateCharts(
	combinedActivity *CombinedCommitActivity,
	grouped bool, mode, stacking, outputPrefix, format string,
	aliases DeveloperAliases,
) error {
	slog.Info("Generating charts", "output_prefix", outputPrefix, "format", format, "mode", mode, "stacking", stacking)

	// Labels for each category
	categories := []struct {
		activityKey func(activity *CommitActivity) map[string][]int
		labels      []string
		filename    string
		title       string
		xLabel      string
	}{
		{func(a *CommitActivity) map[string][]int { return a.Weekdays }, WeekdayLabels(), "by_weekday", "Activity by Weekday", "Weekdays"},
		{func(a *CommitActivity) map[string][]int { return a.Hours }, HourLabels(), "by_hour", "Activity by Hour", "Hours"},
		{func(a *CommitActivity) map[string][]int { return a.Months }, MonthLabels(), "by_month", "Activity by Month", "Months"},
		{func(a *CommitActivity) map[string][]int { return a.Weeks }, WeekLabels(), "by_week", "Activity by Week", "Weeks"},
	}

	for _, category := range categories {
		groupedData := make(map[string]map[string]int)
		xLabel := category.xLabel
		yLabel := "Commits"
		if mode == "lines" {
			yLabel = "Lines of Code"
		}

		switch stacking {
		case "dev", "developer":
			// Group by developer
			for _, repoActivity := range combinedActivity.Repos {
				data := category.activityKey(repoActivity.Activity)
				groupedByDev := prepareGroupedData(data, "dev", repoActivity.RepoName, category.labels)
				mergeGroupedData(groupedData, groupedByDev)
			}

		case "repo":
			// Group by repository
			for _, repoActivity := range combinedActivity.Repos {
				data := category.activityKey(repoActivity.Activity)
				groupedByRepo := prepareGroupedData(data, "repo", repoActivity.RepoName, category.labels)
				mergeGroupedData(groupedData, groupedByRepo)
			}

		default:
			// Flat mode: aggregate everything under a single group
			flatGroup := "All"
			groupedData[flatGroup] = make(map[string]int)
			for _, repoActivity := range combinedActivity.Repos {
				data := category.activityKey(repoActivity.Activity)
				for _, values := range data {
					for i, value := range values {
						groupedData[flatGroup][category.labels[i]] += value
					}
				}
			}
		}

		// Generate chart title and filename
		chartTitle := category.title
		fileName := fmt.Sprintf("%s_%s.%s", outputPrefix, category.filename, format)
		if stacking != "" {
			chartTitle = fmt.Sprintf("%s (%s)", category.title, stacking)
			fileName = fmt.Sprintf("%s_%s_%s.%s", outputPrefix, category.filename, stacking, format)
		}

		// Generate chart for the current category
		err := CreateStackedBarChart(
			groupedData,
			category.labels,
			chartTitle,
			fileName,
			xLabel,
			yLabel,
		)
		if err != nil {
			return fmt.Errorf("error creating chart for %s: %w", category.title, err)
		}
	}

	return nil
}

// CreateStackedBarChart creates a stacked bar chart from the given data
func CreateStackedBarChart(
	data map[string]map[string]int,
	labels []string,
	title, filename, xLabel, yLabel string,
) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	barWidth := vg.Points(20)
	categoryKeys := sortedKeys(data)

	var previousBars *plotter.BarChart
	for j, category := range categoryKeys {
		values := make(plotter.Values, len(labels))
		for i, label := range labels {
			values[i] = float64(data[category][label])
		}

		bars, err := plotter.NewBarChart(values, barWidth)
		if err != nil {
			return fmt.Errorf("could not create bar chart for %s: %w", category, err)
		}

		bars.LineStyle.Width = vg.Length(0)
		bars.Color = colorPalette[j%len(colorPalette)]

		if previousBars != nil {
			bars.StackOn(previousBars)
		}

		p.Add(bars)
		p.Legend.Add(category, bars)

		previousBars = bars
	}

	p.Legend.Top = true
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

func prepareGroupedData(data map[string][]int, groupBy, identifier string, labels []string) map[string]map[string]int {
	groupedData := make(map[string]map[string]int)

	for developer, values := range data {
		key := identifier
		if groupBy == "dev" {
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

func mergeGroupedData(target, source map[string]map[string]int) {
	for key, subMap := range source {
		if target[key] == nil {
			target[key] = make(map[string]int)
		}
		for subKey, value := range subMap {
			target[key][subKey] += value
		}
	}
}
