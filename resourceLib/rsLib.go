package rsLib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs2"
	"github.com/opencontainers/runc/libcontainer/cgroups/manager"
	"github.com/opencontainers/runc/libcontainer/configs"
	"golang.org/x/crypto/ssh/terminal"
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
  Rootless bool
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
    OomKillDisable: true,
	}, nil
}

func IsRootless() bool {
  uid := os.Getuid()
  if uid == 0 {
    return false
  }
  return true
}

func getResourcesRoot(res *Resources) (*cgroup2.Resources, error) {

	var (
		period uint64 = 100000
		weight uint64 = 100
		quota  int64  = int64(res.NumCores) * int64(res.CpuLimit) * 1000
		memMax int64  = int64(res.MaxMem * 1024 * 1024)
	)

	return &cgroup2.Resources{
		CPU: &cgroup2.CPU{
			Weight: &weight,
			Max:    cgroup2.NewCPUMax(&quota, &period),
		},
		Memory: &cgroup2.Memory{
			High: &memMax,
		},
	}, nil
}

func ModifyCgroup(path string, res *Resources) (error) {

  mgr, err := cgroup2.Load(path)
  if err != nil {
    return fmt.Errorf("Unable to load cgroup: %v", err)
  }

  rootRes, err := getResourcesRoot(res)
  if err != nil {
    return fmt.Errorf("Unable to load cgroup: %v", err)
  }

  err = mgr.Update(rootRes)
  if err != nil {
    return fmt.Errorf("Unable to load cgroup: %v", err)
  }

  return nil
}


// Return appropriate manager depending on uid of caller
func CreateManager(res *Resources) (*RsLibHandler, error) {
	rsLib := &RsLibHandler{}
	pid := os.Getpid()
	// If it is in rootless mode retunrn cgroupsmanager
	if res.Rootless {
		groupName := fmt.Sprintf("lmt-%d", os.Getpid())
    uid := os.Getuid()
		cg := &configs.Cgroup{
			Name:     groupName,
			Rootless: res.Rootless,
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

  err = mgr.Update(resConfig)
	if err != nil {
		return rsLib, fmt.Errorf("failed to create a systemd scope: %v", err)
	}

	rootMgr := Cgroups2ManagerRoot{
		mgr: mgr,
	}
	rsLib = &RsLibHandler{
		Mgr:      rootMgr,
		Rootless: res.Rootless,
	}

	return rsLib, nil

}

func PrintPrograms(watch bool) {
	root := "/sys/fs/cgroup"

	for {
    // Get terminal size
    width, err := getTerminalSize()
    if err != nil {
      fmt.Printf("Unable to get terminal size: %v", err)
    }
    colWidth := width / 7
		// Print the header
		headerFormat := fmt.Sprintf("%%-%ds", colWidth)
		fmt.Printf(headerFormat, "MAIN PID")
		fmt.Printf(headerFormat, "CPU USAGE(%)")
		fmt.Printf(headerFormat, "KERNEL MODE(%)")
		fmt.Printf(headerFormat, "USER MODE(%)")
		fmt.Printf(headerFormat, "THROTTLED(%)")
		fmt.Printf(headerFormat, "MEM USAGE(MB)")
		fmt.Printf(headerFormat, "LIMIT(MB)")
		fmt.Println()
		err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() && strings.Contains(path, "lmt") {
				PrintStats(path, width)
			}
			return nil
		})
		if err != nil {
      fmt.Println("Unable to access cgroup directory")
			fmt.Println(err)
		}
    if !watch {
      break
    }
		time.Sleep(2 * time.Second)
		fmt.Print("\033[H\033[2J")
	}
}

func getPidFromPathName(path string) string {
	re := regexp.MustCompile("\\d+")
	return string(re.Find([]byte(path)))
}

func getTerminalSize() (int, error) {
	width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, err
	}
	return width, nil
}

func PrintStats(path string, width int) {
	// Create a new filesystem manager for the given path
	fsmgr, err := fs2.NewManager(nil, path)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the stats for the filesystem manager
	stats, err := fsmgr.GetStats()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Extract the last part of the path to use as the printable path
	shortPath := strings.Split(path, "/")
	cgroupName := shortPath[len(shortPath)-1]

	// Get the process ID for the given printable path
	pid := getPidFromPathName(cgroupName)

	// Calculate the column width based on the given width parameter
	colWidth := width / 7

	// Set up string and float formats for printing the stats
	stringFormat := fmt.Sprintf("%%-%ds", colWidth)
	floatFormat := fmt.Sprintf("%%-%df", colWidth)
	dataFormat := fmt.Sprintf("%%-%dd", colWidth)

	// Get the elapsed time for the process with the given ID
	elapsedTime, err := getProcessElapsedTime(pid)
	if err != nil {
		fmt.Printf("Unable to get process time %v", err)
	}

	// Calculate the CPU usage stats
  totalCpuUsage := float32(stats.CpuStats.CpuUsage.TotalUsage) /
  float32(elapsedTime.Nanoseconds()) * 100 / float32(runtime.NumCPU())
  kernelCpuUsage := float32(stats.CpuStats.CpuUsage.UsageInKernelmode) /
  float32(elapsedTime.Nanoseconds()) * 100 / float32(runtime.NumCPU())
  userCpuUsage := float32(stats.CpuStats.CpuUsage.UsageInUsermode) /
  float32(elapsedTime.Nanoseconds()) * 100 / float32(runtime.NumCPU())
  throttledPercent := float32(stats.CpuStats.ThrottlingData.ThrottledTime) /
  float32(elapsedTime.Nanoseconds()) * 100 / float32(runtime.NumCPU())

	// Calculate the memory usage stats
	memoryUsage := stats.MemoryStats.Usage.Usage / (1024 * 1024)
	memoryLimit := stats.MemoryStats.Usage.Limit / (1024 * 1024)

	// Print the process ID and CPU usage stats
	fmt.Printf(stringFormat, pid)
	fmt.Printf(floatFormat, totalCpuUsage)
	fmt.Printf(floatFormat, kernelCpuUsage)
	fmt.Printf(floatFormat, userCpuUsage)
  fmt.Printf(floatFormat, throttledPercent)

	// Print the memory usage stats
	fmt.Printf(dataFormat, memoryUsage)
	fmt.Printf(dataFormat, memoryLimit)
	fmt.Println()
}

func getProcessElapsedTime(pid string) (time.Duration, error) {
	cmd := exec.Command("ps", "-o", "etime", "-p", pid)
	output, err := cmd.Output()
	lines := strings.Split(string(output), "\n")
	if len(lines) != 3 {
		return 0, fmt.Errorf("invalid output format")
	}

	timeString := strings.TrimSpace(lines[1])
	timeString = strings.ReplaceAll(timeString, ":", "m")
	parsedTime, err := time.ParseDuration(timeString + "s")
	if err != nil {
		return 0, err
	}

	return parsedTime, nil
}

func (rs *RsLibHandler) HandleUnexpectedExits(cmd *exec.Cmd, sigChan chan os.Signal) {
  sig := <-sigChan
  fmt.Println("Got exit signal, closing application: ", sig)
	if rs.Rootless {
		syscall.Kill(-int(cmd.Process.Pid), syscall.SIGKILL)
	}
	rs.Mgr.Delete()
}
