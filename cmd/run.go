/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	// "fmt"

	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	rsLib "github.com/Rohansjamadagni/lmt/resourceLib"
	"github.com/spf13/cobra"
)

var (
	MemLimit float64
	CpuLimit int8
	NumCores int8
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [flags] <program>",
	Short: "Run a command with resource limits",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		res := rsLib.Resources{
			MaxMem:   MemLimit,
			CpuLimit: CpuLimit,
			NumCores: NumCores,
		}
		err := RunCommandWithResources(res, args)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().Float64VarP(&MemLimit, "mem-limit", "m", 0, "Set memory limit in MB")
	runCmd.PersistentFlags().Int8VarP(&CpuLimit, "cpu-limit", "c", 100, "Percentage of cpu to limit the process to")
	runCmd.PersistentFlags().Int8VarP(&NumCores, "num-cores", "n", int8(runtime.NumCPU()), "Number of cores to allow the process to use")
	runCmd.Flags().SetInterspersed(false)
}

func RunCommandWithResources(res rsLib.Resources, args []string) error {
	rsMgr, err := rsLib.CreateManager(&res)
	if err != nil {
		return fmt.Errorf("Error: %w", err)
	}

	cmdName := args[0]
	cmdArgs := args[1:]
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err != nil {
		fmt.Println("Error running command")
		panic(err)
	}

	// Copy the input and output between the terminal window and the command.
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	// Handle exits
	sigChan := make(chan os.Signal)
  signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go rsMgr.HandleUnexpectedExits(cmd, sigChan)
	cmd.Wait()
	return nil
}
