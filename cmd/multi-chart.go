package cmd

import (
	"fmt"
	"log"

	"git-activity/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var multiChartCmd = &cobra.Command{
	Use:   "multi-chart [repos...]",
	Short: "Analyze multiple Git repositories and show them in grouped bar charts",
	Args:  cobra.MinimumNArgs(1), // Requires at least one repository path
	Run: func(cmd *cobra.Command, args []string) {		
		repoPaths := args
		format := viper.GetString("format")
		if format != "png" && format != "svg" {
			log.Fatalf("Invalid format '%s'. Supported formats are 'png' and 'svg'.", format)
		}

		fmt.Printf("Analyzing %d repositories...\n", len(repoPaths))

		// Prepare for grouped bar charts
		repoNames := []string{}
		groupedData := internal.GroupedActivity{}

		// Process each repository
		for _, repoPath := range repoPaths {
			fmt.Printf("Analyzing repository: %s\n", repoPath)
			repoName := internal.GetRepoName(repoPath)

			activity, err := internal.AnalyzeCommits(repoPath)
			if err != nil {
				log.Fatalf("Error analyzing commits for %s: %v", repoPath, err)
			}

			// Add data to grouped activity
			repoNames = append(repoNames, repoName)
			groupedData.Add(repoName, activity)
		}

		// Generate grouped bar charts
		outputPrefix := "multi_repo_chart"

		err := groupedData.CreateGroupedChart("Weekday", groupedData.Weekdays, repoNames, internal.WeekdayLabels(), fmt.Sprintf("%s_by_weekday.%s", outputPrefix, format))
		if err != nil {
			log.Fatalf("Error creating weekday chart: %v", err)
		}
		fmt.Printf("Saved grouped weekday chart as %s_by_weekday.%s\n", outputPrefix, format)

		err = groupedData.CreateGroupedChart("Hour", groupedData.Hours, repoNames, internal.HourLabels(), fmt.Sprintf("%s_by_hour.%s", outputPrefix, format))
		if err != nil {
			log.Fatalf("Error creating hourly chart: %v", err)
		}
		fmt.Printf("Saved grouped hourly chart as %s_by_hour.%s\n", outputPrefix, format)

		err = groupedData.CreateGroupedChart("Month", groupedData.Months, repoNames, internal.MonthLabels(), fmt.Sprintf("%s_by_month.%s", outputPrefix, format))
		if err != nil {
			log.Fatalf("Error creating monthly chart: %v", err)
		}
		fmt.Printf("Saved grouped monthly chart as %s_by_month.%s\n", outputPrefix, format)

		err = groupedData.CreateGroupedChart("Week", groupedData.Weeks, repoNames, internal.WeekLabels(), fmt.Sprintf("%s_by_week.%s", outputPrefix, format))
		if err != nil {
			log.Fatalf("Error creating weekly chart: %v", err)
		}
		fmt.Printf("Saved grouped weekly chart as %s_by_week.%s\n", outputPrefix, format)

		fmt.Println("Multi-repository grouped chart analysis complete.")
	},
}

func init() {
	rootCmd.AddCommand(multiChartCmd)
}
