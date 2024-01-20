package rsLib

import (
	"testing"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/stretchr/testify/assert"
)

func TestGetResourcesRootless(t *testing.T) {
    tests := []struct {
        name     string
        resources Resources
        expected *configs.Resources
    }{
        {
            name: "valid resources",
            resources: Resources{
                MaxMem:  512,
                CpuLimit: 2,
                NumCores: 1,
                Rootless: true,
            },
            expected: &configs.Resources{
                CpuQuota: 2000,
                CpuPeriod: 100000,
                Memory:   536870912,
                OomKillDisable: true,
            },
        },
        // Add more test cases as needed
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual, err := getResourcesRootless(&tt.resources)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, actual)
        })
    }
}


func TestGetResourcesRoot(t *testing.T) {

    var (
      weight uint64 = 100
      cpuQuota int64 = 4000
      cpuPeriod uint64 = 100000
      memory int64 = 1073741824

    )
    tests := []struct {
        name     string
        resources Resources
        expected *cgroup2.Resources
    }{
        {
            name: "valid resources",
            resources: Resources{
                MaxMem:  1024,
                CpuLimit: 2,
                NumCores: 2,
                Rootless: false,
            },
            expected: &cgroup2.Resources{
                CPU: &cgroup2.CPU{
                    Weight: &weight,
                    Max:    cgroup2.NewCPUMax(&cpuQuota, &cpuPeriod),
                },
                Memory: &cgroup2.Memory{
                    High: &memory,
                },
            },
        },
        // Add more test cases as needed
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual, err := getResourcesRoot(&tt.resources)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, actual)
        })
    }
}


