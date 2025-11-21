package nvidia

import (
	"context"
	"fmt"
	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/rpmver"
	"sort"
)

type NvidiaProvider struct {
	PM api.PackageManager
}

var _ api.Provider = (*NvidiaProvider)(nil)

func NewProvider(pm api.PackageManager) *NvidiaProvider {
	return &NvidiaProvider{
		PM: pm,
	}
}

func (i *NvidiaProvider) GetID() string {
	return "nvidia"
}
func (i *NvidiaProvider) GetName() string {
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
	driverPkgs := selectPackagesByNameVersion(all, "nvidia-driver", version, latest)
	cudaPkgs := selectPackagesByNameVersion(all, "nvidia-driver-cuda", version, latest)
	pkgs = append(pkgs, driverPkgs...)
	pkgs = append(pkgs, cudaPkgs...)
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
		"nvidia-fabricmanager",
		"nvidia-fabric-manager-devel",
	}
}

func (i *NvidiaProvider) Install(ctx context.Context, driversInst []api.DriverID) ([]string, error) {
	if i.PM == nil {
		return []string{}, fmt.Errorf("no PackageManager provided for NVIDIA installer")
	}
	driversAvail, err := i.ListAvailable(ctx)
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

	avail, err := i.PM.ListAvailablePackages(ctx)
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

func (i *NvidiaProvider) ListAvailable(ctx context.Context) ([]api.DriverID, error) {
	all, err := i.PM.ListAvailablePackages(ctx)
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
				ProviderID: i.GetID(),
				Version:    pkg.Version,
			})
		}
	}
	sort.Slice(drivers, func(i, j int) bool {
		return rpmver.RpmVersionCompare(drivers[i].Version, drivers[j].Version) > 0
	})
	return drivers, nil
}

func (i *NvidiaProvider) ListInstalled(ctx context.Context) ([]api.DriverID, error) {
	if i.PM == nil {
		return []api.DriverID{}, fmt.Errorf("no PackageManager for NVIDIA installer")
	}

	all, err := i.PM.ListInstalledPackages(ctx)
	if err != nil {
		return []api.DriverID{}, err
	}
	var drivers []api.DriverID
	for _, pkg := range all {
		if pkg.Name == "nvidia-driver" {
			drivers = append(drivers, api.DriverID{
				ProviderID: i.GetID(),
				Version:    pkg.Version,
			})
		}
	}
	sort.Slice(drivers, func(i, j int) bool {
		return rpmver.RpmVersionCompare(drivers[i].Version, drivers[j].Version) > 0
	})
	return drivers, nil
}

func (i *NvidiaProvider) Remove(ctx context.Context, drivers []api.DriverID) ([]string, error) {

	inst, err := i.PM.ListInstalledPackages(ctx)
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

func (i *NvidiaProvider) DetectHardware(ctx context.Context) (bool, error) {
	detector := newAutoDetector()
	return detector.Detect(ctx)
}
