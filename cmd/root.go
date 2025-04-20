package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	fuxxTarget string
	fuxxTool string
)

// rootCmd : default without args
var rootCmd = &cobra.Command{
	Use:   "redi2fuxx",
	Short: "A fuxxing tool for redis.",
	Long:  `A fuxxing tool for redis-based dbms with three mutation modes.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd.PersistentFlags().StringVarP(&fuxxTarget, "target", "t", "redis", "Fuxx Target (redis, keydb, redis-stack)")
	rootCmd.PersistentFlags().StringVarP(&fuxxTool, "tool", "T", "afl", "Fuxx Base (afl, honggfuzz)")
	err := rootCmd.Execute()
	
	if err != nil {
		log.Fatal(err)
	}
	
}