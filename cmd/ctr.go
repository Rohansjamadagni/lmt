/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"runtime"

	rsLib "github.com/Rohansjamadagni/lmt/resourceLib"
  containers "github.com/Rohansjamadagni/lmt/containers"
	"github.com/spf13/cobra"
)

var (
  podman bool
)

// ctrCmd represents the ctr command
var ctrCmd = &cobra.Command{
	Use:   "ctr id",
	Short: "Change the cgroup resource limits of a container",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(1)
		}
		res := rsLib.Resources{
			MaxMem:   MemLimit,
			CpuLimit: CpuLimit,
			NumCores: NumCores,
      Rootless: Rootless,
		}

    if podman {
      err := containers.FindAndSetResPodman(args[0], &res)
      if err != nil {
        fmt.Printf("Error: %v", err)
      }

    } else {
      err := containers.FindAndSetResDocker(args[0], &res)
      if err != nil {
        fmt.Printf("Error: %v", err)
      }
    }

	},
}

func init() {
	setCmd.AddCommand(ctrCmd)
	ctrCmd.PersistentFlags().Float64VarP(&MemLimit, "mem-limit", "m", 0, "Set memory limit in MB")
	ctrCmd.PersistentFlags().Int8VarP(&CpuLimit, "cpu-limit", "c", 100, "Percentage of cpu to limit the process to")
	ctrCmd.PersistentFlags().Int8VarP(&NumCores, "num-cores", "n", int8(runtime.NumCPU()), "Number of cores to allow the process to use")
  ctrCmd.PersistentFlags().BoolVarP(&Rootless, "rootless", "r", rsLib.IsRootless(), "Manually set rootless might be needed inside a container")
  ctrCmd.PersistentFlags().BoolVarP(&podman, "podman", "p", false, "Use with podman containers (default false)")
	runCmd.Flags().SetInterspersed(false)
}
