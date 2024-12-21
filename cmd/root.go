package cmd

import (
	"fmt"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	"github.com/felixge/fgprof"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "git-activity [command]",
	Short: "Analyze Git repository activity",
	Long:  `A CLI tool to analyze commit activity in a Git repository, providing charts for commits by weekday and hour.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior when no subcommands are provided
		fmt.Println("Use 'git-activity analyze --help' to get started.")
	},
}

func Execute() {
	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func MustBind(flagName string, flag *pflag.Flag) {
	if err := viper.BindPFlag(flagName, flag); err != nil {
		log.Fatalf("Error binding %s flag: %v", flagName, err)
	}
}

func init() {
	// Add persistent flags to root command
	rootCmd.PersistentFlags().StringP("start", "s", "", "Start date for analysis (YYYY-MM-DD)")
	rootCmd.PersistentFlags().StringP("end", "e", "", "End date for analysis (YYYY-MM-DD)")
	rootCmd.PersistentFlags().StringP("format", "f", "png", "Output format (png or svg)")
	rootCmd.PersistentFlags().BoolP("grouped", "g", false, "Generate grouped bar charts")
	rootCmd.PersistentFlags().StringP("mode", "m", "commits", "Mode of analysis: 'commits' or 'lines'")
	rootCmd.PersistentFlags().StringP("bars", "b", "", "Stacking mode for bar charts: 'repository', 'developer', or leave empty for flat")
	rootCmd.PersistentFlags().StringP("people", "p", "", "File containing developer aliases")

	// Bind to viper for configuration management using MustBind
	MustBind("start", rootCmd.PersistentFlags().Lookup("start"))
	MustBind("end", rootCmd.PersistentFlags().Lookup("end"))
	MustBind("format", rootCmd.PersistentFlags().Lookup("format"))
	MustBind("grouped", rootCmd.PersistentFlags().Lookup("grouped"))
	MustBind("mode", rootCmd.PersistentFlags().Lookup("mode"))
	MustBind("people", rootCmd.PersistentFlags().Lookup("people"))
	MustBind("bars", rootCmd.PersistentFlags().Lookup("bars"))
}
