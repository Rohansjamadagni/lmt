package containers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	rsLib "github.com/Rohansjamadagni/lmt/resourceLib"
)

// TODO: Find a cleaner easy way to get root
const (
  CGROUP_ROOT = "/sys/fs/cgroup"

  DOCKER_CGROUP_ROOT = CGROUP_ROOT + "/system.slice/"
  DOCKER_PREFIX = "docker-"

  PODMAN_CGROUP_ROOT = CGROUP_ROOT +
    "/user.slice/user-1000.slice/user@1000.service/user.slice/"
  PODMAN_PREFIX = "libpod-"
)

func FindCgroupPath(prefix, baseDir, id string) (string, error) {
	// Ensure the query string is valid.
	if len(id) <= 1 {
		return "", errors.New("id string must be greater than 1")
	}

	// Read the directory contents.
	entries, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return "", err
	}

	var matchingDirs []string
  query := prefix + id

	// Filter entries that match the criteria.
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), query) {
			matchingDirs = append(matchingDirs, entry.Name())
		}
	}

	// Check the number of matching directories.
	switch len(matchingDirs) {

    case 0:
      return "", fmt.Errorf("Unable to find Container with id: %s", id)

    case 1:
      baseDir = strings.TrimPrefix(baseDir, CGROUP_ROOT)
      return filepath.Join(baseDir, matchingDirs[0]), nil

    default:
      return "", fmt.Errorf(
        "Found multiple containers with matching id: %s, please be more specific",
        id)
	}
}

func FindAndSetResDocker(id string, res *rsLib.Resources) (error){

  cgroupPath, err := FindCgroupPath(DOCKER_PREFIX, DOCKER_CGROUP_ROOT, id)

  if err != nil {
    return fmt.Errorf("Unable to set resources: %v", err)
  }

  err = rsLib.ModifyCgroup(cgroupPath, res)
  if err != nil {
    return fmt.Errorf("Unable to set resources: %v", err)
  }

  fmt.Println("Successfully set resources")

  return nil
  
}

func FindAndSetResPodman(id string, res *rsLib.Resources) (error){

  cgroupPath, err := FindCgroupPath(PODMAN_PREFIX, PODMAN_CGROUP_ROOT, id)

  if err != nil {
    return fmt.Errorf("Unable to set resources: %v", err)
  }

  err = rsLib.ModifyCgroup(cgroupPath, res)
  if err != nil {
    return fmt.Errorf("Unable to set resources: %v", err)
  }

  fmt.Println("Successfully set resources")

  return nil
  
}
