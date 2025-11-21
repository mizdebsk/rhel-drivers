package sysinfo

import (
	"bufio"
	"os"
	"github.com/mizdebsk/rhel-drivers/internal/log"
	"runtime"
	"strconv"
	"strings"
)

const osReleasePath = "/etc/os-release"

type SysInfo struct {
	IsRhel    bool
	OsVersion int
	Arch      string
}

func DetectSysInfo() SysInfo {
	arch := detectArch()
	isRhel, osVersion := detectOs(osReleasePath)
	return SysInfo{
		IsRhel:    isRhel,
		OsVersion: osVersion,
		Arch:      arch,
	}
}

func detectArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		// ppc64le, s390x, etc.
		return runtime.GOARCH
	}
}

func detectOs(path string) (bool, int) {
	f, err := os.Open(path)
	if err != nil {
		log.Logf("unable to open %s for reading: %v", path, err)
		return false, 0
	}
	defer f.Close()

	var isRhel bool
	var osVersion int

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ID=") {
			val := strings.TrimPrefix(line, "ID=")
			val = strings.Trim(val, `"`)
			isRhel = val == "rhel"
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			val := strings.TrimPrefix(line, "VERSION_ID=")
			val = strings.Trim(val, `"`)
			if idx := strings.IndexByte(val, '.'); idx >= 0 {
				val = val[:idx]
			}
			n, err := strconv.Atoi(val)
			if err != nil {
				log.Warnf("invalid VERSION_ID %q in %s: %v", val, path, err)
			}
			osVersion = n
		}
	}
	if err := scanner.Err(); err != nil {
		log.Warnf("error parsing %s: %v", path, err)
		return false, 0
	}
	return isRhel, osVersion
}
