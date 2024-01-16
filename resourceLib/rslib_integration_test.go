package rsLib

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCgroupCreationAndDeletion(t *testing.T) {
  // Define the resources to be used for the cgroup
  resources := &Resources{
      MaxMem:   512,
      CpuLimit: 2,
      NumCores: 1,
  }

  if os.Getgid() != 0 {
    // Run the test for rootless modes
    t.Run("Rootless", func(t *testing.T) {

      // Create the cgroup manager
      resources.Rootless = true
      manager, err := CreateManager(resources)
      require.NoError(t, err)
      require.NotNil(t, manager)

      // Get the path of the cgroup
      cgroupPath := manager.Mgr.(CgroupsManagerRootless).mgr.Path("")

      // Check if file exists
      fileInfo, err := os.Stat(cgroupPath)
      fmt.Println(fileInfo.IsDir())
      assert.NoError(t, err, "cgroup directory should exist")

      // Check if pid has been added
      pidValue, err := os.ReadFile(filepath.Join(cgroupPath, "cgroup.procs"))
      assert.NoError(t, err, "Couldn't read cgroup.procs file")
      assert.Equal(t, fmt.Sprintf("%d\n", os.Getpid()), string(pidValue))

      // Check if cpu.max is correct
      cpuMaxValue, err := os.ReadFile(filepath.Join(cgroupPath, "cpu.max"))
      assert.NoError(t, err, "Couldn't read cpu.max file")
      assert.Equal(t, "2000 100000\n", string(cpuMaxValue))

      // Check if memory.max is correct
      memMaxValue, err := os.ReadFile(filepath.Join(cgroupPath, "memory.max"))
      assert.NoError(t, err, "Couldn't read memory.max file")
      assert.Equal(t, "536870912\n", string(memMaxValue))
    })
  } else {

    // TODO: Make test work for root mode

  }

}
