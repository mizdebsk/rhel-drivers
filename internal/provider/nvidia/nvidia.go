package nvidia

import (
	"fmt"
	"sort"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/rpmver"
)

type prov struct {
	PM api.PackageManager
}

var _ api.Provider = (*prov)(nil)

func NewProvider(pm api.PackageManager) api.Provider {
	return &prov{
		PM: pm,
	}
}

func (p *prov) GetID() string {
	return "nvidia"
}
func (p *prov) GetName() string {
	return "NVIDIA"
}

func selectPackagesByNameVersion(all []api.PackageInfo, name, version string, latest bool) []string {
	if latest {
		var best *api.PackageInfo
		for _, pkg := range all {
			if pkg.Name == name && pkg.Version == version {
				if best == nil || rpmver.CompareEVR(best.Epoch, best.Version, best.Release, pkg.Epoch, pkg.Version, pkg.Release) < 0 {
					best = &pkg
				}
			}
		}
		if best == nil {
			return []string{}
		}
		return []string{best.NEVRA()}
	} else {
		var filtered []string
		for _, pkg := range all {
			if pkg.Name == name && pkg.Version == version {
				filtered = append(filtered, pkg.NEVRA())
			}
		}
		return filtered
	}
}

func packageSetVersioned(all []api.PackageInfo, version string, latest bool) []string {
	var pkgs []string
	names := []string{
		"nvidia-driver",
		"nvidia-driver-cuda",
		"nvidia-fabricmanager",
		"nvidia-fabric-manager-devel",
	}
	for _, name := range names {
		selectedPkgs := selectPackagesByNameVersion(all, name, version, latest)
		pkgs = append(pkgs, selectedPkgs...)
	}
	return pkgs
}
func packageSetStatic() []string {
	return []string{
		"cublasmp",
		"cuda-compat",
		"cuda-toolkit",
		"cudnn",
		"dnf-plugin-nvidia",
		"libnccl-devel",
		"libnccl-static",
	}
}

func (p *prov) Install(driversInst []api.DriverID) ([]string, error) {
	if p.PM == nil {
		return []string{}, fmt.Errorf("no PackageManager provided for NVIDIA installer")
	}
	driversAvail, err := p.ListAvailable()
	if err != nil {
		return []string{}, err
	}
	if len(driversAvail) == 0 {
		return []string{}, fmt.Errorf("no NVIDIA driver versions available")
	}
outer:
	for _, driver := range driversInst {
		for _, av := range driversAvail {
			if driver.Version == av.Version {
				continue outer
			}
		}
		return []string{}, fmt.Errorf("no NVIDIA driver version %s available", driver)
	}

	avail, err := p.PM.ListAvailablePackages()
	if err != nil {
		return []string{}, fmt.Errorf("failed to list available packages: %w", err)
	}

	var pkgs []string
	for _, driver := range driversInst {
		pkgs = append(pkgs, packageSetVersioned(avail, driver.Version, true)...)
	}
	pkgs = append(pkgs, packageSetStatic()...)
	return pkgs, nil
}

func (p *prov) ListAvailable() ([]api.DriverID, error) {
	all, err := p.PM.ListAvailablePackages()
	if err != nil {
		return nil, fmt.Errorf("failed to list available packages: %w", err)
	}
	if len(all) == 0 {
		return nil, nil
	}
	var drivers []api.DriverID
	for _, pkg := range all {
		if pkg.Name == "nvidia-driver" {
			drivers = append(drivers, api.DriverID{
				ProviderID: p.GetID(),
				Version:    pkg.Version,
			})
		}
	}
	sort.Slice(drivers, func(i, j int) bool {
		return rpmver.RpmVersionCompare(drivers[i].Version, drivers[j].Version) > 0
	})
	return drivers, nil
}

func (p *prov) ListInstalled() ([]api.DriverID, error) {
	if p.PM == nil {
		return []api.DriverID{}, fmt.Errorf("no PackageManager for NVIDIA installer")
	}

	all, err := p.PM.ListInstalledPackages()
	if err != nil {
		return []api.DriverID{}, err
	}
	var drivers []api.DriverID
	for _, pkg := range all {
		if pkg.Name == "nvidia-driver" {
			drivers = append(drivers, api.DriverID{
				ProviderID: p.GetID(),
				Version:    pkg.Version,
			})
		}
	}
	sort.Slice(drivers, func(i, j int) bool {
		return rpmver.RpmVersionCompare(drivers[i].Version, drivers[j].Version) > 0
	})
	return drivers, nil
}

func (p *prov) Remove(drivers []api.DriverID) ([]string, error) {

	inst, err := p.PM.ListInstalledPackages()
	if err != nil {
		return []string{}, fmt.Errorf("failed to list installed packages: %w", err)
	}

	var pkgs []string
	for _, driver := range drivers {
		pkgs = append(pkgs, packageSetVersioned(inst, driver.Version, false)...)
	}
	pkgs = append(pkgs, packageSetStatic()...)
	return pkgs, nil
}

func (p *prov) DetectHardware() (bool, error) {
	detector := newAutoDetector()
	return detector.Detect()
}
