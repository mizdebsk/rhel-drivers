package amd

import (
	"fmt"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/log"
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

func (p *prov) Install(drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	return []string{"kmod-amdgpu"}, nil
}

func (p *prov) ListInstalled() ([]api.DriverID, error) {
	all, err := p.PM.ListInstalledPackages()
	if err != nil {
		return []api.DriverID{}, err
	}
	var drivers []api.PackageInfo
	for _, pkg := range all {
		if pkg.Name == "kmod-amdgpu" {
			drivers = append(drivers, pkg)
		}
	}
	if len(drivers) == 0 {
		log.Logf("%s driver is currently NOT installed", p.GetName())
		return []api.DriverID{}, nil
	}
	log.Logf("%s driver is currently installed", p.GetName())
	return []api.DriverID{{
		ProviderID: p.GetID(),
		Version:    "latest",
	}}, nil
}

func (p *prov) Remove(drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	return []string{"kmod-amdgpu"}, nil
}

func (p *prov) ListAvailable() ([]api.DriverID, error) {
	all, err := p.PM.ListAvailablePackages()
	if err != nil {
		return []api.DriverID{}, err
	}
	var drivers []api.PackageInfo
	for _, pkg := range all {
		if pkg.Name == "kmod-amdgpu" {
			drivers = append(drivers, pkg)
		}
	}
	if len(drivers) == 0 {
		log.Warnf("%s driver is currently NOT available", p.GetName())
		return []api.DriverID{}, nil
	}
	return []api.DriverID{{
		ProviderID: p.GetID(),
		Version:    "latest",
	}}, nil
}

func (p *prov) DetectHardware() (bool, error) {
	return false, fmt.Errorf("hardware detection for %s is not implemented", p.GetName())
}
