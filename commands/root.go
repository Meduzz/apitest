package commands

import "github.com/spf13/cobra"

var Root *cobra.Command = &cobra.Command{}

func init() {
	Root.Use = "apitest"
	Root.Version = "0.1"
	Root.CompletionOptions.DisableDefaultCmd = true
}
