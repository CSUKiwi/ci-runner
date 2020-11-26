package cmd

import (
	"github.com/spf13/cobra"
	"github.com/fdev-ci/ci-runner/internal/app/run"
)

var cfgFile string

func init() {
	runCmd.Flags().StringVarP(&cfgFile, "config","c","/etc/ci-runner/config.toml", "config file ")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run ci-runner",
	Run: func(cmd *cobra.Command, args []string) {
		run.Run(cfgFile)
	},
}

