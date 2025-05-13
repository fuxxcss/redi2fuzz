package cmd

import (
	"github.com/fuxxcss/redi2fuxx/pkg/fuxx"

	"github.com/spf13/cobra"
)

// fuxxCmd fuxx redis
var fuxxCmd = &cobra.Command {
	Use:   "fuxx",
	Short: "Ready to Fuxx.",
	Run: func(cmd *cobra.Command, args []string) {
		
		fuxx.Fuxx(fuxxTarget, fuxxTool)
	},
}

func init() {

	rootCmd.AddCommand(fuxxCmd)
	
}
