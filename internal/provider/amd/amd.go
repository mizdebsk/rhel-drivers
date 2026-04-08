package amd

import (
	"fmt"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/log"
)

const (
	variantLatest = "latest"

	pkgKmodAmdgpu = "kmod-amdgpu"
	pkgRocm       = "rocm-devel"
)

type prov struct {
	PM api.PackageManager
}

var _ api.Provider = (*prov)(nil)

func (p *prov) GetID() string {
	return "amdgpu"
}
func (p *prov) GetName() string {
	return "AMD GPU"
}

func NewProvider(pm api.PackageManager) api.Provider {
	return &prov{
		PM: pm,
	}
}

func combinedPackages() []string {
	return []string{pkgKmodAmdgpu, pkgRocm}
}

func hasAnyPackage(all []api.PackageInfo, names ...string) bool {
	for _, name := range names {
		if hasPackage(all, name) {
			return true
		}
	}
	return false
}

func hasAllPackages(all []api.PackageInfo, names ...string) bool {
	for _, name := range names {
		if !hasPackage(all, name) {
			return false
		}
	}
	return true
}

func validateDrivers(drivers []api.DriverID, providerName string) error {
	for _, d := range drivers {
		if d.Version != variantLatest {
			return fmt.Errorf("unknown %s driver variant: %s", providerName, d.Version)
		}
	}
	return nil
}

func hasPackage(all []api.PackageInfo, name string) bool {
	for _, pkg := range all {
		if pkg.Name == name {
			return true
		}
	}
	return false
}

func (p *prov) Install(drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	if err := validateDrivers(drivers, p.GetName()); err != nil {
		return nil, err
	}
	return combinedPackages(), nil
}

func (p *prov) ListInstalled() ([]api.DriverID, error) {
	all, err := p.PM.ListInstalledPackages()
	if err != nil {
		return []api.DriverID{}, err
	}
	packages := combinedPackages()
	if !hasAnyPackage(all, packages...) {
		log.Logf("%s drivers are currently NOT installed", p.GetName())
		return []api.DriverID{}, nil
	}
	if !hasAllPackages(all, packages...) {
		log.Warnf("%s drivers are partially installed", p.GetName())
		return []api.DriverID{{ProviderID: p.GetID(), Version: variantLatest}}, nil
	}
	log.Logf("%s drivers are currently installed", p.GetName())
	return []api.DriverID{{ProviderID: p.GetID(), Version: variantLatest}}, nil
}

func (p *prov) Remove(drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	if err := validateDrivers(drivers, p.GetName()); err != nil {
		return nil, err
	}
	return combinedPackages(), nil
}

func (p *prov) ListAvailable() ([]api.DriverID, error) {
	all, err := p.PM.ListAvailablePackages()
	if err != nil {
		return []api.DriverID{}, err
	}
	if !hasAllPackages(all, combinedPackages()...) {
		log.Warnf("%s drivers are currently NOT available", p.GetName())
		return []api.DriverID{}, nil
	}

	return []api.DriverID{{ProviderID: p.GetID(), Version: variantLatest}}, nil
}

func (p *prov) DetectHardware() (bool, error) {
	detector := newAutoDetector()
	return detector.Detect()
}
