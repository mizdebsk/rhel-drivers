package dnf

import (
	"fmt"
	"strings"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/cache"
	"github.com/mizdebsk/rhel-drivers/internal/log"
)

const defaultDNFBinary = "dnf"

type pkgMgr struct {
	bin  string
	exec api.Executor
}

var _ api.PackageManager = (*pkgMgr)(nil)

func NewPackageManager(executor api.Executor) api.PackageManager {
	return &pkgMgr{
		bin:  defaultDNFBinary,
		exec: executor,
	}
}

var availableCache = cache.Cache[[]api.PackageInfo]{}
var installedCache = cache.Cache[[]api.PackageInfo]{}

func (pm *pkgMgr) ListAvailablePackages() ([]api.PackageInfo, error) {
	return availableCache.Get(func() ([]api.PackageInfo, error) {
		tags := []string{"name", "epoch", "version", "release", "arch", "sourcerpm", "repoid"}
		// QQQ and YYY are there to make filtering spurious lines easier.
		format := "QQQ"
		for _, field := range tags {
			format += "|%{" + field + "}"
		}
		// Trailing NL is not required with DNF 4, but will be required with DNF 5.
		// With DNF 4 it will result in empty lines, but they are ignored anyway.
		format += "|YYY\n"
		lines, err := pm.exec.RunCapture(pm.bin, []string{"-q", "repoquery", "--qf", format}...)
		if err != nil {
			return nil, fmt.Errorf("failed to list available packages: %w", err)
		}
		return parseQueryOutput(lines), nil
	})
}

func (pm *pkgMgr) ListInstalledPackages() ([]api.PackageInfo, error) {
	return installedCache.Get(func() ([]api.PackageInfo, error) {
		tags := []string{"NAME", "EPOCH", "VERSION", "RELEASE", "ARCH", "SOURCERPM"}
		format := "QQQ"
		for _, field := range tags {
			format += "|%|" + field + "?{%{" + field + "}}|"
		}
		format += "||YYY\n"
		lines, err := pm.exec.RunCapture("rpm", []string{"-qa", "--qf", format}...)
		if err != nil {
			return nil, fmt.Errorf("failed to list installed packages: %w", err)
		}
		return parseQueryOutput(lines), nil
	})
}

func parseQueryOutput(lines []string) []api.PackageInfo {
	var infos []api.PackageInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "QQQ|") && strings.HasSuffix(line, "|YYY") {
			line = strings.TrimPrefix(line, "QQQ|")
			line = strings.TrimSuffix(line, "|YYY")
			fields := strings.SplitN(line, "|", 7)
			if len(fields) == 7 {
				infos = append(infos, api.PackageInfo{
					Name:       fields[0],
					Epoch:      fields[1],
					Version:    fields[2],
					Release:    fields[3],
					Arch:       fields[4],
					SourceName: parseNameFromNVRA(fields[5]),
					Repo:       fields[6],
				})
			}
		}
	}
	return infos
}

func (pm *pkgMgr) Install(packages []string) error {
	if len(packages) == 0 {
		log.Warnf("no packages to install")
		return nil
	}
	log.Logf("installing packages: %v", packages)
	return pm.exec.Run(pm.bin, append([]string{"install"}, packages...))
}

func (pm *pkgMgr) Remove(packages []string) error {
	if len(packages) == 0 {
		log.Warnf("no packages to remove")
		return nil
	}
	log.Logf("removing packages: %v", packages)
	return pm.exec.Run(pm.bin, append([]string{"remove"}, packages...))
}
