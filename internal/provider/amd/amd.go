package amd

import (
	"context"
	"fmt"

	"github.com/mizdebsk/rhel-drivers/internal/api"
	"github.com/mizdebsk/rhel-drivers/internal/log"
)

type AmdProvider struct {
	PM api.PackageManager
}

var _ api.Provider = (*AmdProvider)(nil)

func (i *AmdProvider) GetID() string {
	return "amdgpu"
}
func (i *AmdProvider) GetName() string {
	return "AMD GPU"
}

func NewProvider(pm api.PackageManager) *AmdProvider {
	return &AmdProvider{
		PM: pm,
	}
}

func (i *AmdProvider) Install(ctx context.Context, drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	return []string{"kmod-amdgpu"}, nil
}

func (i *AmdProvider) ListInstalled(ctx context.Context) ([]api.DriverID, error) {
	all, err := i.PM.ListInstalledPackages(ctx)
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
		log.Logf("%s driver is currently NOT installed", i.GetName())
		return []api.DriverID{}, nil
	}
	log.Logf("%s driver is currently installed", i.GetName())
	return []api.DriverID{{
		ProviderID: i.GetID(),
		Version:    "latest",
	}}, nil
}

func (i *AmdProvider) Remove(ctx context.Context, drivers []api.DriverID) ([]string, error) {
	if len(drivers) == 0 {
		return []string{}, nil
	}
	return []string{"kmod-amdgpu"}, nil
}

func (i *AmdProvider) ListAvailable(ctx context.Context) ([]api.DriverID, error) {
	all, err := i.PM.ListAvailablePackages(ctx)
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
		log.Warnf("%s driver is currently NOT available", i.GetName())
		return []api.DriverID{}, nil
	}
	return []api.DriverID{{
		ProviderID: i.GetID(),
		Version:    "latest",
	}}, nil
}

func (i *AmdProvider) DetectHardware(ctx context.Context) (bool, error) {
	return false, fmt.Errorf("hardware detection for %s is not implemented", i.GetName())
}
