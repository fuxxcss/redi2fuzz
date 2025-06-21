package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/fuxxcss/redi2fuzz/pkg/utils"
)

var (
	fuzzTarget utils.TargetType
)

// rootCmd : default without args
var rootCmd = &cobra.Command{
	Use:   "redi2fuzz",
	Short: "A fuzzing tool for redis.",
	Long:  `A fuzzing tool for redis-based dbms with graph mutation mode.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	var arg string
	rootCmd.PersistentFlags().StringVarP(&arg, "target", "t", "redis", "Fuzz Target (redis, keydb, redis-stack)")

	switch arg {
	case "redis", "Redis" :
		fuzzTarget = utils.REDI_REDIS
	case "keydb", "KeyDB" :
		fuzzTarget = utils.REDI_KEYDB
	case "redis-stack", "Redis Stack", "Redis-Stack":
		fuzzTarget = utils.REDI_STACK
	default:
		log.Fatalf("err: %s is not support\n", fuzzTarget)
	}

	err := rootCmd.Execute()
	
	if err != nil {
		log.Fatal(err)
	}
	
}