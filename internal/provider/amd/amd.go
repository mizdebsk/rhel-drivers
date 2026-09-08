package amd

import (
	"fmt"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/hwdetect"
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

func stackPackages() []string {
	return []string{pkgKmodAmdgpu, pkgRocm}
}

func partitionPackages(all []api.PackageInfo, names ...string) (present, missing []string) {
	for _, name := range names {
		if hasPackage(all, name) {
			present = append(present, name)
		} else {
			missing = append(missing, name)
		}
	}
	return present, missing
}

func hasPackage(all []api.PackageInfo, name string) bool {
	for _, pkg := range all {
		if pkg.Name == name {
			return true
		}
	}
	return false
}

func validateDrivers(drivers []api.DriverID, providerName string) error {
	for _, driver := range drivers {
		if driver.Version != variantLatest {
			return fmt.Errorf("unknown %s driver variant: %s", providerName, driver.Version)
		}
	}
	return nil
}

func (p *prov) Install(drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	if err := validateDrivers(drivers, p.GetName()); err != nil {
		return nil, err
	}
	return stackPackages(), nil
}

func (p *prov) ListInstalled() ([]api.DriverID, error) {
	all, err := p.PM.ListInstalledPackages()
	if err != nil {
		return []api.DriverID{}, err
	}
	packages := stackPackages()
	installed, missing := partitionPackages(all, packages...)
	switch len(missing) {
	case len(packages):
		log.Logf("%s stack is currently NOT installed", p.GetName())
		return []api.DriverID{}, nil
	case 0:
		log.Logf("%s stack is currently installed", p.GetName())
	default:
		log.Warnf("%s stack is partially installed; installed: %s; missing: %s",
			p.GetName(), strings.Join(installed, ", "), strings.Join(missing, ", "))
	}
	return []api.DriverID{{ProviderID: p.GetID(), Version: variantLatest}}, nil
}

func (p *prov) Remove(drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	if err := validateDrivers(drivers, p.GetName()); err != nil {
		return nil, err
	}
	return stackPackages(), nil
}

func (p *prov) ListAvailable() ([]api.DriverID, error) {
	all, err := p.PM.ListAvailablePackages()
	if err != nil {
		return []api.DriverID{}, err
	}
	_, missing := partitionPackages(all, stackPackages()...)
	if len(missing) > 0 {
		log.Warnf("%s stack is currently NOT available; missing: %s", p.GetName(), strings.Join(missing, ", "))
		return []api.DriverID{}, nil
	}
	return []api.DriverID{{ProviderID: p.GetID(), Version: variantLatest}}, nil
}

func (p *prov) DetectHardware() (bool, error) {
	return detectHardware(hwdetect.DefaultModaliasRoot)
}
