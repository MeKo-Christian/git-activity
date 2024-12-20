package cmd

import (
	"fmt"
	"log"

	"git-activity/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [repos...]",
	Short: "Analyze multiple Git repositories and combine their data",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Retrieve flag values
		startStr := viper.GetString("start")
		endStr := viper.GetString("end")
		format := viper.GetString("format")
		grouped := viper.GetBool("grouped")
		mode := viper.GetString("mode")

		// Parse dates
		start, end := parseDateRange(startStr, endStr)

		// Validate format
		if format != "png" && format != "svg" {
			log.Fatalf("Invalid format '%s'. Supported formats are 'png' and 'svg'.", format)
		}

		// Validate mode
		if mode != "commits" && mode != "lines" {
			log.Fatalf("Invalid mode '%s'. Supported modes are 'commits' or 'lines'.", mode)
		}

		// Perform analysis
		outputPrefix, combinedActivity := internal.AnalyzeRepositories(args, start, end, mode)

		if grouped {
			// Generate grouped bar charts
			err := internal.GenerateGroupedCharts(combinedActivity, mode, outputPrefix, format)
			if err != nil {
				log.Fatalf("Error generating charts: %v", err)
			}
		} else {
			// Generate standard charts
			err := internal.GenerateCharts(combinedActivity, mode, outputPrefix, format)
			if err != nil {
				log.Fatalf("Error generating charts: %v", err)
			}
		}

		fmt.Println("Repository analysis complete.")
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
