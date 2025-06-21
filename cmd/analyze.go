package cmd

import (

	"github.com/fuxxcss/redi2fuzz/pkg/analyze"

	"github.com/spf13/cobra"
)

// analyzeCmd analyze bugs
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze Bugs.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		analyze.Analyze(fuzzTarget, path)
	},
}

func init() {

	rootCmd.AddCommand(analyzeCmd)
	
}