package cmd

import "github.com/spf13/cobra"

var (
	url string
)

func init() {
	registerCmd.Flags().StringVarP(&url, "url", "u", "", "apiserver address")

	rootCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "register a new runner",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

