package cmd

import (
	"github.com/fuxxcss/redis-fuxx/pkg/"

	"github.com/spf13/cobra"
)

// targetCmd fuxx target dbms
var targetCmd = &cobra.Command{
	Use:   "target",
	Short: "Fuxx Target (redis, keydb, redis-stack)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		db.StartUp(target)
	},
}

func init() {

	rootCmd.AddCommand(targetCmd)
	
}
