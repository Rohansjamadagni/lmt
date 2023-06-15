# lmt
lmt is a cli program written in go that can be used to run applications with resource limits enforced using cgroupsv2 

# Requirements

This program works only on linux and requires cgroupsv2 support. Most newer linux distros already come with cgroupsv2 installed and enabled by default.

The following distributions are known to use cgroup v2 by default:

  - Fedora (since 31)
  - Arch Linux (since April 2021)
  - openSUSE Tumbleweed (since c. 2021)
  - Debian GNU/Linux (since 11)
  - Ubuntu (since 21.10)
  - RHEL and RHEL-like distributions (since 9)

On other systemd-based distros, cgroup v2 can be enabled by adding `systemd.unified_cgroup_hierarchy=1` to the kernel cmdline.

You can check if its enabled and mounted using the following command:
```
mount | grep -q 'cgroup2' && echo "cgroupsv2 hierarchy is mounted" || echo "cgroupsv2 hierarchy is NOT mounted"
```

# Installation
Make sure go is installed
```
go install "github.com/Rohansjamadagni/lmt@latest"
```
# Usage

## Run
```
Usage:
  lmt run [flags] <program>

Flags:
  -c, --cpu-limit int8    Percentage of cpu to limit the process to [0-100] (default 100)
  -h, --help              help for run
  -m, --mem-limit float   Set memory limit in MB (default 0 [no limit])
  -n, --num-cores int8    Number of cores to allow the process to use (default [nproc])
```

## PS
``` 
Usage:
  lmt ps [flags]

Flags:
  -h, --help    help for ps
  -w, --watch   watch command output
```

# Examples

## Run command
```
lmt run -m 500 firefox
```
Note: Interactive TUI apps are a bit tricky to handle, so as a workaround launch a tmux session (or even a new terminal session) with the desired limits and run applications from there.

## PS command

Gives a live view of the resources used by programs run with lmt.

```
lmt ps
lmt ps -w
```
