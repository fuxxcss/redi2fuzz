package cmd

import (
	"github.com/spf13/cobra"
	
	"github.com/fuxxcss/redi2fuzz/pkg/fuzz"
)

// fuzzCmd 
var fuzzCmd = &cobra.Command {
	Use:   "fuzz",
	Short: "Ready to Fuzz.",
	Run: func(cmd *cobra.Command, args []string) {

		fuzz.Fuzz(fuzzTarget)
	},
}

func init() {

	rootCmd.AddCommand(fuzzCmd)
	
}
