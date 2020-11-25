package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tluo-github/ci-runner/internal/pkg/common"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ci-runner",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version:" + common.VERSION)
	},
}
