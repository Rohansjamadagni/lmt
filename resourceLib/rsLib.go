package rsLib

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/manager"
	"github.com/opencontainers/runc/libcontainer/configs"
)

type Manager interface {
	Delete()
}

// For rootless
type CgroupsManagerRootless struct {
	mgr cgroups.Manager
}

// For root
type Cgroups2ManagerRoot struct {
	mgr *cgroup2.Manager
}

type RsLibHandler struct {
	Mgr      Manager
	Rootless bool
}

type Resources struct {
	MaxMem   float64
	CpuLimit int8
	NumCores int8
}

func (cm Cgroups2ManagerRoot) Delete() {
	cm.mgr.Delete()
	cm.mgr.Kill()
	os.Exit(1)
}

func (cm CgroupsManagerRootless) Delete() {
	mgr := cm.mgr
	os.Remove(mgr.Path(""))
	os.Exit(1)
}

func getResourcesRootless(res *Resources) (*configs.Resources, error) {
	return &configs.Resources{
		CpuQuota:  int64(res.NumCores) * int64(res.CpuLimit) * 1000,
		CpuPeriod: 100000,
		Memory:    int64(res.MaxMem * 1024 * 1024),
	}, nil
}

func getResourcesRoot(res *Resources) (*cgroup2.Resources, error) {

	var (
		period uint64 = 100000
		weight uint64 = 100
		quota  int64  = int64(res.NumCores) * int64(res.CpuLimit)
		memMax int64  = int64(res.MaxMem * 1024 * 1024)
	)

	return &cgroup2.Resources{
		CPU: &cgroup2.CPU{
			Weight: &weight,
			Max:    cgroup2.NewCPUMax(&quota, &period),
		},
		Memory: &cgroup2.Memory{
			Max: &memMax,
		},
	}, nil
}

// Return appropriate manager depending on uid of caller
func CreateManager(res *Resources) (*RsLibHandler, error) {
	rsLib := &RsLibHandler{}
	pid := os.Getpid()
	// If it is in rootless mode retunrn cgroupsmanager
	if uid := os.Getuid(); uid != 0 {
		groupName := fmt.Sprintf("lmt-%d", os.Getpid())
		cg := &configs.Cgroup{
			Name:     groupName,
			Rootless: true,
			OwnerUID: &uid,
			Systemd:  true,
		}

		// Initiate a new manager
		mgr, err := manager.New(cg)
		if err != nil {
			return rsLib, fmt.Errorf("failed to init new cgroup manager: %v", err)
		}

		// Add the pid to the cgroup
		err = mgr.Apply(pid)
		if err != nil {
			return rsLib, fmt.Errorf("failed to apply pid: %v", err)
		}

		// set the resources
		resConfig, err := getResourcesRootless(res)
		err = mgr.Set(resConfig)
		if err != nil {
			return rsLib, fmt.Errorf("failed to set resoure limits: %v", err)
		}

		rootlessMgr := CgroupsManagerRootless{
			mgr: mgr,
		}
		rsLib = &RsLibHandler{
			Mgr:      rootlessMgr,
			Rootless: true,
		}
		return rsLib, nil
	}

	// In root mode create resources and manager
	resConfig, err := getResourcesRoot(res)
	if err != nil {
		return rsLib, fmt.Errorf("failed to get resources: %v", err)
	}

	groupPath := fmt.Sprintf("lmt-%d.scope", pid)
	mgr, err := cgroup2.NewSystemd("", groupPath, pid, resConfig)
	if err != nil {
		return rsLib, fmt.Errorf("failed to create a systemd scope: %v", err)
	}

	rootMgr := Cgroups2ManagerRoot{
		mgr: mgr,
	}
	rsLib = &RsLibHandler{
		Mgr:      rootMgr,
		Rootless: false,
	}

	return rsLib, nil

}

func (rs *RsLibHandler) HandleUnexpectedExits(cmd *exec.Cmd, sigChan chan os.Signal) {
	<-sigChan
	if rs.Rootless {
		syscall.Kill(-int(cmd.Process.Pid), syscall.SIGKILL)
	}
	rs.Mgr.Delete()
}
