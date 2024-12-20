package internal

import (
	"fmt"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// GroupedActivity tracks data for multiple repositories
type GroupedActivity struct {
	Weekdays map[string][]int
	Hours    map[string][]int
	Months   map[string][]int
	Weeks    map[string][]int
}

// Add data for a specific repository
func (ga *GroupedActivity) Add(repoName string, activity *CommitActivity) {
	if ga.Weekdays == nil {
		ga.Weekdays = make(map[string][]int)
		ga.Hours = make(map[string][]int)
		ga.Months = make(map[string][]int)
		ga.Weeks = make(map[string][]int)
	}

	ga.Weekdays[repoName] = activity.Weekdays[:]
	ga.Hours[repoName] = activity.Hours[:]
	ga.Months[repoName] = activity.Months[:]
	ga.Weeks[repoName] = activity.Weeks[:]
}

// CreateGroupedChart generates a grouped bar chart
func (ga *GroupedActivity) CreateGroupedChart(chartTitle string, data map[string][]int, repoNames []string, labels []string, filename string) error {
	// Prepare plot
	p := plot.New()
	p.Title.Text = fmt.Sprintf("Normalized Commits by %s", chartTitle)
	p.Y.Label.Text = "Proportion"
	p.NominalX(labels...)

	// Calculate spacing and bar width
	totalGroups := len(repoNames)
	groupWidth := vg.Length(35)                           // Total width allocated per tick (including space)
	barWidth := groupWidth / vg.Length(totalGroups) * 0.8 // Overlap slightly
	barSpacing := vg.Length(5)                            // Space between ticks

	// Ensure minimum bar width of 3px
	if barWidth < vg.Points(3) {
		barWidth = vg.Points(3)
	}

	// Adjust total group width if bars are small
	if totalGroups > 1 && groupWidth > barWidth*vg.Length(totalGroups) {
		groupWidth = barWidth*vg.Length(totalGroups) + vg.Points(1)
	}

	// Add grouped bars
	for i, repoName := range repoNames {
		// Normalize values
		values := data[repoName]
		total := sum(values)
		normalizedValues := make([]float64, len(values))
		if total > 0 {
			for j, value := range values {
				normalizedValues[j] = float64(value) / float64(total)
			}
		}

		// Create bar chart for normalized values
		bars, err := plotter.NewBarChart(plotter.Values(normalizedValues), barWidth)
		if err != nil {
			return fmt.Errorf("could not create bar chart for %s: %w", repoName, err)
		}

		bars.LineStyle.Width = vg.Length(0)
		bars.Color = colorPalette[i%len(colorPalette)] // Cycle through color palette

		// Offset bars within each tick, allow overlap
		bars.Offset = (vg.Length(i)-vg.Length(totalGroups)/2)*barWidth + groupWidth/2 - barWidth*vg.Length(totalGroups)/2

		p.Add(bars)
	}

	// Adjust spacing between ticks
	p.X.Padding = barSpacing

	// Save chart
	err := p.Save(14*vg.Inch, 6*vg.Inch, filename)
	if err != nil {
		return fmt.Errorf("could not save chart: %w", err)
	}

	return nil
}

// Utility method to sum values in a slice
func sum(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

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
