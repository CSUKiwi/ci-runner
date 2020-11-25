package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)


func init() {

}
var rootCmd = &cobra.Command{
	Use:   "ci-runner",
	Short: "ci-runner",
	Run: func(cmd *cobra.Command, args []string) {
		Error(cmd, args, errors.New("unrecognized command"))
	},
}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
