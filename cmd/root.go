package cmd

import (
	"github.com/spf13/cobra"
	"squad-utils/cmd/agent"
)

var rootCmd = &cobra.Command{
	Use:   "sq-utils",
	Short: "给你的战术臭队加点料",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(agent.Cmd)
}
func Execute() {
	rootCmd.Execute()
}
