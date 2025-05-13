package cmd

import (
	"github.com/fuxxcss/redi2fuxx/pkg/analyze"

	"github.com/spf13/cobra"
)


// analyzeCmd analyze bugs
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze Bugs.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bug := args[0]
		analyze.Analyze(fuxxTarget, bug)
	},
}

func init() {

	rootCmd.AddCommand(analyzeCmd)
	
}