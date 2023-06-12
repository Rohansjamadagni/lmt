/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	rsLib "github.com/Rohansjamadagni/lmt/resourceLib"
	"github.com/spf13/cobra"
)

// psCmd represents the list command
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List the running processes created by lmt and their usage",
	Run: func(cmd *cobra.Command, args []string) {
		ListProcesses()
	},
}

var watch bool

func init() {
	rootCmd.AddCommand(psCmd)
	psCmd.PersistentFlags().BoolVarP(&watch, "watch", "w", false, "watch command output")
}

func ListProcesses() {
	rsLib.PrintPrograms(watch)
}
