package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	fuxxMode string
	fuxxTarget string
)

// rootCmd : default without args
rootCmd := &cobra.Command{
	Use:   "redi2fuxx",
	Short: "A fuxxing tool for redis.",
	Long:  `A tool used to fuxx redis-based dbms with gramfree mutation.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd.PersistentFlags().StringVarP(&fuxxMode, "mode", "m", "dumb", "Fuxx Mode (dumb, gramfree, fagent)")
	rootCmd.PersistentFlags().StringVarP(&fuxxMode, "target", "t", "redis", "Fuxx Target (redis, keydb, redis-stack)")
	err := rootCmd.Execute()
	
	if err != nil {
		log.Fatal(err)
	}
	
}