# lmt
lmt is a cli program written in go that can be used to run applications with resource limits enforced using cgroupsv2 

# Installation
clone the repo and run:
```
go get .
go build
```
# Usage
```
Usage:
  lmt run [flags] <program>

Flags:
  -c, --cpu-limit int8    Percentage of cpu to limit the process to [0-100] (default 100)
  -h, --help              help for run
  -m, --mem-limit float   Set memory limit in MB (default 0 [no limit])
  -n, --num-cores int8    Number of cores to allow the process to use (default [nproc])
```

# Examples

```
lmt run -m 500 firefox
```
