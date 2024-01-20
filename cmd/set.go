package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [flags] id",
	Short: "Set resource limits on container or a pid",
	Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Invalid Command: ", args)
    fmt.Println("Usage: lmt set ctr <flags> <id>")
    os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
