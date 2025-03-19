package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var fuxxMode string

// rootCmd : default without args
rootCmd := &cobra.Command{
	Use:   "redis-fuxx",
	Short: "A fuxxing tool for redis.",
	Long:  `A tool used to fuxx redis-based dbms with gramfree mutation.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd.PersistentFlags().StringVarP(&fuxxMode, "mode", "m", "dumb", "Fuxx Mode (dumb, gramfree)")
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("err: %v\n",err)
		os.Exit(1)
	}
	
}